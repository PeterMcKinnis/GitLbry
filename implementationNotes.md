

# Implementation Notes

## General Implementation Strategy

Git [remote-helpers](https://git-scm.com/docs/gitremote-helpers) are programs that allow git to sync with an arbitrary data store.  `git-lbry` is an implementation of a remote helper for the lbry protocol.  The remote helper protocol allows for arbitrary storage and retrieval of git objects (blobs, trees, commits, and tags), it is quite flexible but there is a bunch of detail that seems trickey to get right.  The `remote-helpers` protocol communicates over stdin and stdout and there are  more or less three different sub-protocols: 
1. connect
2. push/fetch
3. fast-import, fast-export

Connect seems to be the most modern and feature rich.  It allows reduction of network bandwitch by omiting unwanted or duplicate data and supports fetching limited commit histories.  It also seemes to be the hardest to implement.

Push/Fetch is a very simple but open-ended protocol.  Git gives the remote-helper a list of tags/branches to fetch and the remote-helpter updates the files in the local .git folder directly.  For a psuh, git gives the remote-helper a list of tags/branches and remote-helper stores them however it chooses.

FastImport / Fast Export: This is designed for use with git fast-import and git fast-export commands.  Documentation seemed poor.

It was chosen to use the Push/Fetch option.  A full clone of the remote repo is maintained locally.  When pushing, `git bundle` is used to create a patch (e.g. a diff from the remote to local).  When fetching, git downloads all patches from the lbry network, applies them local clone using `git bundle unbundle` then imports them to the actual git repo using `git fetch-pack`

### Downsides of current implementation

1. There is a full local copy of the remote repo.  This can confuse some editors and tools.  I may want to tinker with the path and storage location to confuse fewer tools.   A more complex implemenation could probably be faster and use less disk space, but meh, both of those are pretty cheap these days.

### Upsides of current implementation

1. Each `git push` generates a single patch file on the lbry network.  This makes resolution of concurrant pushes easy, and is optimized for the lbry network which seems to be designed for a small number of larger files.
2. The implementation is as ignorant as possible about inner workings of git.  The implementation has no notion of object, tree, blob, or commit storage, reference depths, .git directory structure, .binary file formats, tags, commits, branches, etc.    

## Authorized Push Users

git-lbry allows teams of authorized users to push a git repo hosted on the lbry network.  

We authorize users by adding an entry to the file at the root of the repo. (e.g. `lbry://org_name-repo_name`)  This file maintains a list of lbry channels with push authorization and the patch_index range where their pushes are authorized.  See root.json format for details.


## Local Directory Structure

Directory Tree

```  
  some_source_file.c
  another_source_file.c
  .git
  .glbry
    <repohash>.lock               A file system lock for the <repohash> folder
    <repohash>                                
      .git                        A full local copy of the lbry version of the repo
      root.json                   Controlls permissions, etc
      info.json                   House keeping information, see format below
      <n>-<hash_prev_commit>      Local copy of a bundle file downloaded from the lbry
                                  network.  Uses git's native .bundle format
      out
        <n>-<hash_prev_commit>    Local copy of a bundle file being uploaded to the lbry
                                  network.  Uses git's native .bundle format.
```

## Repo Hash Generation

`<repohash>` is the sha1 hash of the url to the repo as used by git remote.  Hash gives a repo a unique name on disk and avoids characters that may not be allowed for file paths in certain environments.  Sha1 was used for convieniance because that's what git uses for everything and it was needed anyway for the project.

## Lbry Directory Structure

```
lbry://org.com|project    - Controlls Permissions
lbry://org.com|project-0  - Patch 0
lbry://org.com|project-1  - Patch 1
...
lbry://org.com|project-227  - Patch 227
```

## Lbry Patch Conflicts

Gitlbry stores repo data on the lbry network as a list of patches to the repo.  Each patch is assigned a monitonically increasing patch_index starting from zero.  Each patch (except for the zero-th) also includes the hash of prior patch on which it is based.

Due to the public and decentralized nature of the network it is possible for there to be multiple canidateds for each patch.  E.g. we may have several different versions of patch 221 to choose from. In this case `gitlbry` uses the following is used to resolve the cannonical patch.

1. If patch has been marked deleted, it is non-cannonical.  See Root Json section for details.

2. If patch was createded by a user (e.g. wallet) without push permissions, it is non-cannonical.  See Root Json section for details 

3. All patches that are based on a non-cannonical prior patch are also non-cannonical.  Not applicable to the 0th patch.

4. The oldest patch is cannonical. The lbry **depth?** for the patch file is used to determine age

## File Format (info.json)

info.json includes the index and sha1 has of the most recently imported cannonical patch.  Noteably the hash is for current patch (will be used as the prior patch when we try to push)

example

```
{
  patch: {
    idx: 221
    sha1: "ccdddd6c5b19436e52146dfc11fd8632ca60b31b",
  }
}
```

## File Format (Repo Root)

The repo root contains permissions for who is allowed to push to the repo and when they are allowed to push.  A patch is valid if:
2. The channel that published the patch is listed in the users and contains at least one range such that start <= patch_index && (patch_index < end || end == -1)

Deleted contains a list of files in the repo that are deleted.  lbry only allows a file owner to delete / modify a file.  This provision allows the repo owner to simulate deletion by maintaining a list of deleted files which will then be ignored by the tooling.

```
{

  gitlbry: 1  // version number.  Only version 1 is supported
  authors: [
    {
      claim_id: <string>      // e.g. "e66aa0b46d98caf5aeafcee0bbb89bdafec0de72"
      channel_name: <string>  // e.g. "@gitlbry" must start with "@"
      ranges: [
        {
          start: <int>        // Seconds from unix epoch inclusive. -1 indicates no lower bound
          end: <int>          // Seconds from unix epoch exclusive. -1 indicates no upper bound
        }
      ]
    }
  ]
  deleted: [<string>]   // ID's for claims that will be ignored
}
```

# Program Outline


Start-Up

  1. Aquire local filesystem lock at `.glbry/<repohash>/`

  2. Ensure that `lbry://repo/url` points to a valid gitlbry repo, if not, fail.

  2. Initialize local repo clone if necessary

  3. Load info.json

  4. Update .gitlbry/<reposhash>/in from the lbry network
    1. Set patch id based on info.json
    2. Find first patch name with commit-depth >= 1 `lbry://repo/url/n-<priorhash>`
    3. If none exists, done
    4. download patch
    5. Apply locally
    6. update info.json
    7. inc patchid and repeat at step 2

    5. (done) Apply each bundle to .git repo at
      .gitlbry/<repohash>/.git


Capabilities
  Wrie to std out, super simple, just supporting List, Push, and Fetch

List:
  1. Get all refs by running git list-refs as .gitlbry/<repohash>/.git
  2. Parse inputs and save for later

  3. Optionally write out head
  4. Write out regular references

Push:
  1. (done) Read all push lines from stdin - stop at blank line
  2. (done) Attemp to push locally to file://.gitlbry/<repohash>/ 
  3. Bundle Necessary Objects using git-bundle
    Exclude hashes from #2
    Include hashes from #1
    Get Name 
    Save copy locally to .gitlbry/<repohash>/out/<n>-<hash_prev_patch>
    Send to lbry network
      -> use resolve to search for potential conflicts anywhere in mem-pool
      -> fail if there are conflicts 
      -> send bundle to lbry block untill it makes it into the mem-pool
  4. (done) Write to stdin \n indicating we completed our work

Fetch:
  1. Not much to do here since lbry was synced at starup.  
  2. Read Fetch line from stdin
  3. forward to file://.gitlbry/<repohash>/.git 
  4. Output appropriate response

Teardown:
  4. Filesystem unlock at .gitlbry/<repohash>/


# Support for multiple users

## Design goals

Have a simple interface to add, remove, create users.  

1. I think the easiest way will be have the user 
2. Only the owner of t

## Lbry Malice and Corruption

There are many ways to vandalize or corrupt git lbry if someone gets publish access to the channel or an author become hostile.  

1. Author deletes some cannonical patches that author published.
  - Ruins the repo ... need to fix this at some point 🤔
2. Author posts sensitive data into the repo
  - Repo owner cannot perminatly delete lbry file files from other user's accounts and it will remain publically accessable.  Passwords and Credentials will need to be changed ASAP.
2. Author accidently posts a very large file
  - owner can fix by pseudo-deleting the file
  - readers download and ignore wasting some time.
3. Author posts a corrupt patch
  - owner can fix by pseudo-deleting the file otherwise
  - readers download and ignore wasting some time.
4. Author posts a malicious patch that introduces bugs / vulnerabilities into the software
  - This is a management concern but not unique to git-lbry
  - owner can fix by pseudo-deleting the file if
5. User persistantly pushes
  - current implementation fails when another user has a patch in the lbrymem pool, so this allows an author to to easily crate a DOS attack.  Solution is to ban the author.
6. User spams thousands of garbage files
  - This is fine, they will just be ignored by tooling

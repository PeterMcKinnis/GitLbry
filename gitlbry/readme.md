
# Command - Init

## Useage

`lit init <lbry_url>`

## Summary

Creates an empty repo at the given `<lbry_url>` and prints the perminant url to it.

`<lbry_url>`  - The lbry url for the repo.  You may omit the "lbry://" prefix for convieniance. 



# Command - Author

## Useage

`gitlbry author <lbry_url> [[^]<channel_url>]*`

## Summary

With zero channel_urls, prints a list all all channels with current or prior push permssions and their current permission.  Each line has the format

```
@<channel_name>#<channel_id> (granted | revoked)
```

With one or more `<channel_url>` grants and revokes push privilidges for each channel.   Channels prefixed with ^ will have push privlidge revoked, others will have push privledge granted.  Error if any channel cannot be resolved on the lbry network 

`<lbry_url>`  A lbry url to the repository.  For convieniance, the prefix "lbry://" may be omitted.

`<channel_url>` - The lbry url for the channel.  For convieniance, the prefix "lbry://" or "lbry://@" may be omitted.   e.g. the following are all permitted "lbry://@my_channel", "@my_channel", "my_channel", "lbry://my_channel:2a34fe", "my_channel:2a34fe", etc.

Example - Give Permission to `lbry://@HalbertReeves` and `lbry://@ArnoldGarrett:f2` and revokes privilidges for `lbry://HughRussell`
```
gitlbry author-modify my.org-fastsort @HalbertReeves ArnoldGarrett:f2 ^HughRussell
```

Example - List privilidges
```
gitlbry author-modify my.org-fastsort
```



# Command - me

## Useage

`gitlbry me [<channel_url>]`

## Summary

With zero arguements, prints the perminant lbry for the channel being used to push changes

With one arguement, sets the channel to be used to push future changes

`<channel_url>` - The lbry url for the channel.  For convieniance, the prefix "lbry://" or "lbry://@" may be omitted.   e.g. the following are all permitted "lbry://@my_channel", "@my_channel", "my_channel", "lbry://my_channel:2a34fe", "my_channel:2a34fe", etc.










# Just a Stub, Not Implemented conflicts

`gitlbry conflicts [--min=<n>] [--max=<n>]`

## Summary

Outputs a list of all patches with conflicts (see below for details).  Shows the winning patch, and reason that the other patches are loosing.

Gitlbry stores repo data on the lbry network as a list of patches to the repo.  These patches are numbered 0, 1, 2, etc.  Due to the decentralized nature of the network it is possible that multiple users can push same patch.  E.g. we may have two different versions of patch 221 on the lbry network creating a patch conflict on patch 221. In this case lbry uses the following is used to resolve the cannonical patch.

1. Patches that are part of a non-canonical branches are ignored.  e.g. if a user pushes 2 patches around the same time another user pushes a patch.  Both of the new patches may become part of a non-cannoical branch and will be ignored.

2. Manual Selection - In rare cases you may want to manually select the correct patch id.  See `gitlbry fix` for details (tbd)

3. Older patches have precidence over newer patches.  The lbry sequence specifier for the patch file is used to determine precidence


## Output format

```
Patch <n>
$<SequenceId> ("winning" | ("loosing" ("non-cannonical-branch" | "older" | "manually-disabled"))
```

## Example output

```
Patch 220
$1 patch winning
$2 loosing older
Patch 221
$1 loosing non-cannnonical-branch
$2 winning
```

## Options

**--min=\<n>** exclude conflicts where patch_id is less than `n`  N is a non-negative decimal integer.  Default is to include all results

**--max\<n>** exclude conflicts where patch_id is greater than or equal to `n`.  `n` must be formatted as a decimal integer.  Default is to include all results


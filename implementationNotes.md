


To generate repo hash use:
  1. Start with the full lbry url e.g.
    lbry://@mychannel$1/myrepo
  2. Apply unicode normilization according to UnicodeNormilization Form D (NFD)
  3. Apply lowercase using en_us local
  4. Take the sha1 hash for the result
  5. convert to an lcase hex string e.g "ccdddd6c5b19436e52146dfc11fd8632ca60b31b"


in/info.json
  {
    bundleIndex: 1  // The number of bundles that have been sucessfully applied to the .git repo
  }


Start-Up

  1. Filesystem lock at .gitlbry/<repohash>/

  2. (done) Initialize if necessary

  3. (done) Load info.json

  4. (done) Update .gitlbry/<reposhash>/in from the lbry network
    -> For dev just copy from /out

  5. (done) Apply each bundle to .git repo at
    .gitlbry/<repohash>/.git

  6. (done) Update info.json

Capabilities
  Wrie to std out, super simple, just supporting List, Push, and Fetch

List:
  1. Get all refs by running git list-refs as .gitlbry/<repohash>/.git
  2. Parse inputs and save for later

  3. Optionally write out head
  4. Write out regular references

Push:
  1. (done) Read all push lines until blank line
  2. (done) Attemp to push locally to file://.gitlbry/<repohash>/.git 
    -> Respond to stdin as appropriate for each push
  3. Pack Necessary Objects
    Exclude hashes from #2
    Include hashes from #1 
    Save to:
  4. (done) Write to stdin \n indicating we completed our work

Fetch:
  1. Read Fetch line
    -> forward to file://.gitlbry/<repohash>/.git 
    -> output appropriate response
  2. Output /n to indicate completion

Teardown:
  4. Filesystem unlock at .gitlbry/<repohash>/

Directory Tree
  
  .git
  .gitlbry/<repohash>.lock -> used a as file system lock for the <repohash> folder
  .gitlbry/<repohash>/
    .git  -> This a full local copy of the lbry version of the repo
    in/info.json -> sync info for remotes - see info.json for format
    in/<n>.bundle -> bundles recently downloaded from lbry but not applied locally .git yet.  This uses git's native bundle format.
    out/<n>.bundlee -> bundles in process of being pushed to lbry


































Depricated


LBRY Bundle
  JSON Header
  {
    version: 1
    refs: [
      {
        name: "/ref/heads/master"
        value: "ccdddd6c5b19436e52146dfc11fd8632ca60b31b" || "" (empty string to delete)
      },
      {
        name: "/ref/heads/feature5"
        value: "46dfc11fd8632ca60b31bccdddd6c5b19436e521"
      },
      {
        name: "/ref/heads/master"
        value: "" (empty string to delete)
      },
    ]
    head: "ccdddd6c5b19436e52146dfc11fd8632ca60b31b" or "/ref/heads/master"
  }
  Consists of a json formatted header followed by the contents of .pack


1. Todo:
  update cli to match docs
    1. Need fxn to resolve channel
    2. Need fxn to resolve new_url
    3. Need fxn to resolve existing_url
    4. Update init to work with or without a channel

2. Review Push and update as necessary



Issues:
1. Grant / Revoke / Init
  - Need to do a claim search and fail if there is stuff that is really recent in the mem-pool

2. Need "Permissions" function that does bulk grant / revoke

3. Re-read lbry and ensure all sdk calls are blocking

3. Misc updated to get git push / pulls working
  a. Edits in push.go
    - 1. Check that there aren't others trying to push
    - 2. Add code to stream_create to lbry network
  b. Edits in fetch.go (maybe no though...)
    Nothing to change
  c. Edits in startup.go
    - 1. Reload permissions file if necessary
    - 2. In Loop - Download patch n
  d. Local directory structure has changed a bit
4. Read / review / delete dead code in lbry.go and cli.go


# Do some preliminary tests


## Buy lbc ... need debit card :(




package main

// Aquire a file system lock for the files in .git/<repohash>
func (rh RepoHash) lock() error {
	// Todo implement
	// lock contention which can be ignored in early dev.
	return nil;
}

// Releases a file system lock for the files in .git/<repohash>
func (rh RepoHash) unlock() error {
	// Todo implement
	// lock contention which can be ignored in early dev.
	return nil;
}
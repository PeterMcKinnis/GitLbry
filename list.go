package main


func (s Startup) list() error {
	Printf("%v %v\n", s.head, "HEAD");
	return s.listForPush();
}

func (s Startup) listForPush() error {
	for _, x := range s.refs {
		Printf("%v %v\n", x.ref.toHexString(), x.name);
	}
	Printf("\n");
	return nil;
}



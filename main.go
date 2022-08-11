package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

// TODO this shouldn't be hard coded



var stdinReader = bufio.NewReader(os.Stdin);
var pushArgRe = regexp.MustCompile(`^push (\+?)([^:]*):([^:]*)$`);

type pushArg struct {
	force bool
	localRef string
	remoteRef string
}

func readPushInputs() ([]pushArg, error) {

	var result []pushArg;
	for {
		text, err := stdinReader.ReadString('\n')
		if (err != nil) {
			return nil, err;
		}

		if (text == "") {
			break;
		}

		matches := pushArgRe.FindStringSubmatch(text);
		result = append(result, pushArg{
			force: matches[1] == "+",
			localRef: matches[2],
			remoteRef: matches[3],
		})
	}

	return result, nil;
}

func parsePushCmd(cmd string) (pushArg, error) {

	matches := pushArgRe.FindStringSubmatch(cmd);
	return pushArg{
		force: matches[1] == "+",
		localRef: matches[2],
		remoteRef: matches[3],
	}, nil;
}

func getLocalObjType(sha string) (string, error) {

	var args = []string {
		"cat-file",
		"-t",
		sha,
	};

  out, err := exec.Command(
    "git", args...,
  ).Output();

	if err != nil {
		return "", err;
	}

	return strings.Trim(string(out), "\r\n"), nil;

}

func getLocalObjContent(sha string) ([]byte, error) {

	// printErr(fmt.Sprintf("getLocalObjContent sha %v \n", sha));

	var prefix = sha[:2];
	var suffex = sha[2:];

	// Make Directory if necessary
	filePath := path.Join(".git", "objects", prefix, suffex)

	result, err := ioutil.ReadFile(filePath);

	if err != nil {
		printErr("Error reading content\n")
	}

	return result, err;
}

func saveContentToRemote(sha string, url string, content []byte) error {

	var prefix = sha[:2];
	var suffex = sha[2:];

	var url = url[7:];

	// Make Directory if necessary
	dirPath := path.Join(base_dir, path, prefix)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err;
	}

	filePath := path.Join(dirPath, suffex);
	return ioutil.WriteFile(filePath, content, 0666);
}

func pushObject(sha string) error {

	if (sha == "") {
		return nil;
	}

	// Get Type of object
	content, err := getLocalObjContent(sha);

	if err != nil {
		printErr("147");
		return err;
	}

	err = saveContentToRemote(sha, content);
	if err != nil {
		printErr("153")
	}

	return err;

}

func push(url string, cmd string) (er error) {

	// Parse Command
	arg, err := parsePushCmd(cmd);
	if err != nil {
		return err;
	}

	defer func ()  {
		if er != nil {
			print(fmt.Sprintf("error %v %v\n", arg.remoteRef, er));
		}
	}();

	// Get SHA1 Id for all Remote Refs
	refs, err := listRemoteRefSha();
	if err != nil {
		return err;
	}

	// Use git command rev-list to find all
	// objects that need to be added to remote
	var args = []string {
		"rev-list",
		"--objects",
		"--no-object-names",
		arg.localRef,
	};

	for _, ref := range refs {
		args = append(args, "^" + ref);
	}

  out, err := exec.Command(
    "git", args...,
  ).Output();

	if err != nil {
		return err;
	}

	// Just for debugging
	// fmt.Fprint(os.Stderr, string(out));

	objList :=	strings.Split(string(out), "\n");


	// Upload all objects
	for _, obj := range objList {
		err := pushObject(obj);
		if err != nil {
			return err;
		}
	}

	// TODO save ref
	print(fmt.Sprintf("ok %v\n", arg.remoteRef));
	return nil;
}

/// This gets the sha1 hash (as a string) for every branch and tag
func listRemoteRefSha() ([]string, error) {
	
	var result []string;

	// Add Heads
	heads := path.Join(remote_dir, "refs", "heads");
	entries, err := os.ReadDir(heads);
	if err != nil {
		return nil, err;
	}

	for _, entry := range entries {
		path := path.Join(heads, entry.Name());
		content, err := ioutil.ReadFile(path);
		if (err != nil) {
			return nil, err;
		}
		contentStr := strings.TrimRight(string(content), "\n");
		result = append(result, contentStr);
	}

		// Add Heads
		tags := path.Join(remote_dir, "refs", "tags");
		entries, err = os.ReadDir(tags);
		if err != nil {
			return nil, err;
		}
	
		for _, entry := range entries {
			path := path.Join(heads, entry.Name());
			content, err := ioutil.ReadFile(path);
			if (err != nil) {
				return nil, err;
			}
			contentStr := strings.TrimRight(string(content), "\n");
			result = append(result, contentStr);
		}
	
	return result, nil;
}

func list() error {

	remote_dir := "C:\\Users\\peter\\git-remote-lbry\\data\\remote-lbry";
	
	// Add Heads
	heads := path.Join(remote_dir, "refs", "heads");
	entries, err := os.ReadDir(heads);
	if err != nil {
		return nil;
	}

	for _, entry := range entries {
		path := path.Join(heads, entry.Name());
		content, err := ioutil.ReadFile(path);
		contentStr := strings.TrimRight(string(content), "\n");
		if err != nil {
			return err;
		}
		print(fmt.Sprintf("%v refs/head/%v\n",  contentStr, entry.Name()));
	}

	// Add Tags
	tags := path.Join(remote_dir, "refs", "tags");
	entries, err = os.ReadDir(tags);
	if err != nil {
		return nil;
	}

	for _, entry := range entries {
		path := path.Join(heads, entry.Name());
		content, err := ioutil.ReadFile(path);
		contentStr := strings.TrimRight(string(content), "\n");
		if err != nil {
			return err;
		}
		print(fmt.Sprintf("%v refs/tags/%v\n",  contentStr, entry.Name()));
	}

	// Todo use head file for remote head...
	print("@refs/heads/master HEAD\n");

	print("\n");
	return nil;
}


func readLine() (string, error) {
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Echo to std out for debugging
	fmt.Fprint(os.Stderr, "<< " + line);
	return line, nil;
}

func print(str string) {
	// Echo to std out for debugging
	printInfo(str);
	fmt.Print(str);
}

func printErr(str string) {
	fmt.Fprint(os.Stderr, "err >> " + str);
}

func printInfo(str string) {
	fmt.Fprint(os.Stderr, ">> " + str);
}

func Main() (er error) {
  if len(os.Args) > 3 {
    return fmt.Errorf("Usage: git-remote-lbry remote-name url")
  }



  remoteName := os.Args[1]
  url := os.Args[2]
	printInfo(fmt.Sprintf("args, %v", os.Args));

	// Add "path" to the import list
	// localdir := path.Join(os.Getenv("GIT_DIR"), "go", remoteName)

	for {
		// Note that command will include the trailing newline.
		command, err := readLine()
		if err != nil {
			return err
		}
		command = strings.TrimRight(command, "\r\n");

		switch  {
		case strings.HasPrefix(command, "capabilities"):
			print("fetch\n")
			print("push\n")
			print("\n")
		case strings.HasPrefix(command, "list"):
			/*
			refs, err := GitListRefs()
			if err != nil {
				return fmt.Errorf("command list: %v", err)
			}
		
			head, err := GitSymbolicRef("HEAD")
			if err != nil {
				return fmt.Errorf("command list: %v", err)
			}

			for refname := range refs {
				fmt.Printf("? %s\n", refname)
			}
	
			fmt.Printf("@%s HEAD\n", head)
		*/

			// No refs present until we finish push
			//fmt.Fprint(os.Stderr, "listing not implemented");
			//return errors.New("list not implemented");
			//print("8fa6b3625bc9541acb6f104ea06260c2a2c49ea0 refs/heads/master\n");
			//print("@refs/heads/master HEAD\n");
			list();

		case strings.HasPrefix(command, "fetch"):
			log.Fatalf("not implemented");

		case strings.HasPrefix(command, "push"):
			push(remoteName, url, command);
		case command == "":
			return nil
		default:
			return fmt.Errorf("Received unknown command %q", command)
		}
	}
}

func main() {
  if err := Main(); err != nil {
    log.Fatal(err)
  }
}



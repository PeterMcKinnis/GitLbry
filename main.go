package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)


type sha1 int 

type tree struct {
	id sha1
	children []treeChild
}

type treeChild struct {
	id	sha1

	/// False if blob
	isTree bool
	name string
}

type blob struct {
	id sha1
	contents []byte
}

type commit struct {
	id sha1
	tree sha1
	previous *sha1
}

type tag struct {
	object sha1;
}

type gitObject interface {
	nested() []sha1
}


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

func push(remoteName string, url string, cmd string) error{

	// Parse Command
	arg, err := parsePushCmd(cmd);
	if err != nil {
		return err;
	}

	// Get All Remote Refs
	//refs, err := listRemotes();
	//if err != nil {
	//	return err;
	//}

	var args = []string {
		"rev-list",
		"--objects",
		arg.localRef,
	};

	//for _, rr := range refs {
	//	args = append(args, "^" + rr);
	//}

  out, err := exec.Command(
    "git", args...,
  ).Output();

	if err != nil {
		return err;
	}

	log.Print(string(out));

	// Parse All Objects
	return errors.New("Not implemented");
	
}


func list() error {
	remote_dir := "C:\\Users\\peter\\git-remote-lbry\\data\\remote-lbry";
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

	// Todo also list remote tags

	// Todo use head file for remote head...
	print("@refs/heads/master HEAD\n");

	print("\n");
	return nil;
}

/*
func GitListRefs() (map[string]string, error) {
  out, err := exec.Command(
    "git", "for-each-ref", "--format=%(objectname) %(refname)",
    "refs/heads/",
  ).Output()
  if err != nil {
    return nil, err
  }

  lines := bytes.Split(out, []byte{'\n'})
  refs := make(map[string]string, len(lines))

  for _, line := range lines {
    fields := bytes.Split(line, []byte{' '})

    if len(fields) < 2 {
      break
    }

    refs[string(fields[1])] = string(fields[0])
  }

  return refs, nil
}

func GitSymbolicRef(name string) (string, error) {
  out, err := exec.Command("git", "symbolic-ref", name).Output()
  if err != nil {
    return "", fmt.Errorf(
      "GitSymbolicRef: git symbolic-ref %s: %v", name, out, err)
  }

  return string(bytes.TrimSpace(out)), nil
}
*/

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
	fmt.Fprint(os.Stderr, ">> " + str);
	fmt.Print(str);
}


func Main() (er error) {
	

  if len(os.Args) > 3 {
    return fmt.Errorf("Usage: git-remote-lbry remote-name url")
  }

  remoteName := os.Args[1]
  url := os.Args[2]
  
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
			log.Fatalf("not implemented");

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



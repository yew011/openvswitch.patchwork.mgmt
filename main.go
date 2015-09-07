package main

import (
       "bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"os"
	"os/exec"
	"regexp"
)

const (
	FLAG_OVS_DIR             = "ovs-dir"
	FLAG_OVS_COMMIT          = "ovs-commit"
)

func main() {
	app := cli.NewApp()
	app.Name = "ovs-patchwork"
	app.Usage = "Tool for ovs patchwork facilitation. " +
		    "User must provide the --ovs-dir and --ovs-commit options."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   FLAG_OVS_DIR,
			Usage:  "Path to ovs git repo",
		},
		cli.StringFlag{
			Name:   FLAG_OVS_COMMIT,
			Usage:  "Protocol used to connect mysql server",
		},
	}
	app.Writer = os.Stdout
	log.SetOutput(os.Stdout)
	app.Before = func(c *cli.Context) error {
		return nil
	}
	app.Action = func(c *cli.Context) {
		var duplicates []string
		// maps between patch name to the non-duplicate 'pwclient list' line
		patches := make(map[string]string)

		cmd := exec.Command("python", "./pwclient", "list", "-s", "NEW")
		cmd_stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("'pwclient list' cannot get stdout pipe: %s", err)
		}
		err = cmd.Start()
		if err != nil {
			log.Fatalf("'pwclient list' run error: %s", err)
		}
		scanner := bufio.NewScanner(cmd_stdout)
		// 'pwclient list' output has format "ID STATE    NAME"
		re := regexp.MustCompile(`^[0-9]+  New\s+\[.*\] (.*)$`)
		for scanner.Scan() {
			if submatch := re.FindStringSubmatch(scanner.Text()); submatch != nil {
				// classifies the patches
				if _, ok := patches[re.FindStringSubmatch(scanner.Text())[1]]; ok {
					duplicates = append(duplicates, scanner.Text())
				} else {
					patches[re.FindStringSubmatch(scanner.Text())[1]] = scanner.Text()
				}
			}
		}
		if scanner.Err() != nil {
			log.Fatalf("'pwclient list' scanner error: %s", scanner.Err())
		}
		err = cmd.Wait()
		if err != nil {
			log.Fatalf("'pwclient list' exit error: %s", err)
		}
		if duplicates != nil {
			fmt.Println("Duplicate Patches in Patchwork")
			fmt.Println("==============================")
			fmt.Println("ID      State        Name")
			fmt.Println("--      -----        ----")
			for _, dup := range duplicates {
				fmt.Println(dup)
			}
		}
	}
	app.Run(os.Args)
}

package main

import (
       "bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"os"
	"os/exec"
	"regexp"
	"time"
)

const (
	FLAG_OVS_DIR             = "ovs-dir"
	FLAG_OVS_COMMIT          = "ovs-commit"
	FLAG_MARK_COMMITTED      = "mark-committed"
	FLAG_MARK_DUP            = "mark-dup"
)

type Pair struct {
	ID      string
	line    string
}

const dateShortForm = "2006-01-02"
const stateACCEPT   = "Accepted"
const stateDUP      = "Not Applicable"

var OVSDIR, OVSCOMMIT string

/*
 * Dumps all 'NEW' patch entries from patchwork.  Reports duplicated
 * entries, records the rest of entries in a string->string map and
 * returns it.
 */
func do_duplication_check() (map[string]Pair, []string) {
	var duplicates, duplicates_ID, outdated []string
	/* maps between patch name to the non-duplicate 'pwclient list' line. */
	patches := make(map[string]Pair)

	today := time.Now()

	/* checks for outdated and duplicated patch records. */
	cmd := exec.Command("python", "./pwclient", "list", "-s", "NEW",
			    "-f", "%{id}  %{state}    %{date}   %{name}")
	cmd_stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("'pwclient list' cannot get stdout pipe: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalf("'pwclient list' run error: %s", err)
	}
	scanner := bufio.NewScanner(cmd_stdout)
	/* 'pwclient list' output has format "ID  STATE   DATE   NAME".
	 * extracts "DATE" and "NAME". */
	re := regexp.MustCompile(`^([0-9]+)  New    ([-0-9]+) .*   \[.*\] (.*)$`)
	for scanner.Scan() {
		if submatch := re.FindStringSubmatch(scanner.Text());
		   submatch != nil {
			idField     := submatch[1]
			dateField   := submatch[2]
			commitField := submatch[3]

			/* cherry-picks more than 30-day old patches. */
			date, err := time.Parse(dateShortForm, dateField)
			if err != nil {
				log.Fatalf("'pwclient list' cannot parse " +
					   "date: %s", err)
			}
			if today.Sub(date) / (24 * time.Hour) > 30 {
				outdated = append(outdated, scanner.Text())
			}

			/* classifies the patches.  since the oldest record
			 * comes first, we just kick it out of the map when
			 * hitting a collison. */
			if _, ok := patches[commitField]; ok {
				duplicates = append(duplicates,
						    patches[commitField].line)
				duplicates_ID = append(duplicates_ID,
						       patches[commitField].ID)
			}
			patches[commitField] = Pair{idField, scanner.Text()}
		}
	}
	if scanner.Err() != nil {
		log.Fatalf("'pwclient list' scanner error: %s", scanner.Err())
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("'pwclient list' exit error: %s", err)
	}

	if outdated != nil {
		fmt.Println("30+ Day Old Patches")
		fmt.Println("===================")
		fmt.Println("ID      State  Date                  Name")
		fmt.Println("--      -----  ----                  ----")
		for _, out := range outdated {
			fmt.Println(out)
		}
		fmt.Println()
	}

	if duplicates != nil {
		fmt.Println("Duplicate Patches in Patchwork")
		fmt.Println("==============================")
		fmt.Println("ID      State  Date                  Name")
		fmt.Println("--      -----  ----                  ----")
		for _, dup := range duplicates {
			fmt.Println(dup)
		}
		fmt.Println()
	}

	return patches, duplicates_ID
}

/*
 * Dumps specified commit history from ovs repo.  If the commit is
 * found in "patches", it means that the patch has already been
 * upstreamed and thusly should be marked as "accepted".
 */
func do_committed_check(patches map[string]Pair) []string {
	var committed, committed_ID []string

	cmd := exec.Command("git", "log", "--oneline", "-n", "500", OVSCOMMIT)
	cmd_stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("'git log --oneline' cannot get stdout pipe: %s", err)
	}
	cmd.Dir = OVSDIR
	err = cmd.Start()
	if err != nil {
		log.Fatalf("'git log --oneline' run error: %s", err)
	}
	scanner := bufio.NewScanner(cmd_stdout)
	/* 'git log --oneline' output has format "ID NAME". */
	re := regexp.MustCompile(`^[0-9a-f]+ (.*)$`)
	for scanner.Scan() {
		if submatch := re.FindStringSubmatch(scanner.Text()); submatch != nil {
			/* if the committed patch name is found in 'patches',
			 * record the 'pwclient list' entry in 'committed'. */
			if elem, ok := patches[submatch[1]]; ok {
				committed = append(committed, elem.line)
				committed_ID = append(committed_ID, elem.ID)
			}
		}
	}
	if scanner.Err() != nil {
		log.Fatalf("'git log --oneline' scanner error: %s", scanner.Err())
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("'git log --oneline' exit error: %s", err)
	}
	if committed != nil {
		fmt.Println("Committed Patches in Patchwork")
		fmt.Println("==============================")
		fmt.Println("ID      State  Date                  Name")
		fmt.Println("--      -----  ----                  ----")
		for _, entry := range committed {
			fmt.Println(entry)
		}
		fmt.Println()
	}

	return committed_ID
}

/*
 * Given array of patch 'ids' in patchwork, mark those patches
 * to 'state'.
 */
func do_state_update(ids []string, state string) {
	args := append([]string{"./pwclient", "update", "-s", state}, ids...)
	cmd := exec.Command("python", args...)
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("'pwclient update' run error: %s", err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("'pwclient update' exit error: %s", err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name  = "ovs-patchwork"
	app.Usage = "Tool for ovs patchwork facilitation.  " +
		    "User must provide the --ovs-dir and --ovs-commit options."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   FLAG_OVS_DIR,
			Usage:  "Path to ovs git repo",
		},
		cli.StringFlag{
			Name:   FLAG_OVS_COMMIT,
			Usage:  "Commit to start check for committed patches",
		},
		cli.BoolFlag{
			Name:   FLAG_MARK_COMMITTED,
			Usage:  "Mark the committed patch as 'Accepted'",
		},
		cli.BoolFlag{
			Name:   FLAG_MARK_DUP,
			Usage:  "Mark the duplicate patch as 'Not Applicable'",
		},
	}
	app.Writer = os.Stdout
	app.Action = func(c *cli.Context) {
		if OVSDIR = c.String(FLAG_OVS_DIR); OVSDIR == "" {
			log.Fatalf("must provide option --ovs-dir")
		}
		if OVSCOMMIT = c.String(FLAG_OVS_COMMIT); OVSCOMMIT == "" {
			log.Fatalf("must provide option --ovs-commit")
		}
		patches, duplicates_ID := do_duplication_check()
		committed_ID := do_committed_check(patches)
		if c.Bool(FLAG_MARK_COMMITTED) {
			fmt.Println("Mark Committed as Accepted")
			fmt.Println("==========================")
			do_state_update(committed_ID, stateACCEPT)
		}
		if c.Bool(FLAG_MARK_DUP) {
			fmt.Println("Mark Dup as Not Applicable")
			fmt.Println("==========================")
			do_state_update(duplicates_ID, stateDUP)
		}
	}
	app.Run(os.Args)
	log.SetOutput(os.Stdout)
}

package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"./plane"
)

// initLogger binds logger to log file, returning the file to which the logger is writing.
func initLogger(radius int) *os.File {
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		os.Mkdir("logs", os.ModeDir)
	}

	logfile, _ := os.OpenFile("logs/"+strconv.Itoa(radius)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, fs.ModePerm)
	log.SetOutput(logfile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	return logfile
}

// parseArgs parses and returns following command-line agruments:
// -r       - radius of the quarter
// -workers - amount of concurrent workers
// -timeout - time in minutes, after which program stops calculations prematurely
func parseArgs() (uint, uint, uint) {
	var r, workers, timeout uint

	flag.UintVar(&r, "r", 0, "radius of the quarter")
	flag.UintVar(&workers, "workers", 1, "amount of concurrent workers")
	flag.UintVar(&timeout, "timeout", 0, "time in minutes, after which program stops calculations prematurely")

	flag.Parse()
	return r, workers, timeout
}

// saveResults writes results to a file, in human-readable form.
func saveResults(r int, weights []uint64) {
	resfile, err := os.OpenFile("result_"+strconv.Itoa(r)+".dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: Couldn't open file to save results to: %v\n", err)
	}
	defer resfile.Close()

	var buffer bytes.Buffer
	for key := 1; key <= r; key++ {
		buffer.WriteString(fmt.Sprintf("%d:%d\n", key, weights[key]))
	}

	writer := bufio.NewWriter(resfile)
	if _, err = writer.Write(buffer.Bytes()); err != nil {
		log.Fatalf("ERROR: Couldn't write data to save file: %v\n", err)
	}
	writer.Flush()
}

func main() {
	r, workers, timeout := parseArgs()

	initLogger(int(r))
	log.Printf("INFO: Parsed flags: r: %d; workers: %d; timeout: %d\n", r, workers, timeout)

	login()

	pln := plane.InitNewPlane(r, workers)

	if err := pln.TryToRestoreState(); err != nil {
		log.Printf("INFO: Couldn't restore state of last execution: %v\n", err)
	}

	pln.Start()

	if timeout != 0 {
		select {
		case <-pln.NotifyOnFinish():
			break
		case <-time.After(time.Duration(timeout) * time.Minute):
			pln.Stop()
			break
		}
	} else {
		<-pln.NotifyOnFinish()
	}

	log.Printf("INFO: Calculations have stopped, finished state: %t\n", pln.IsFinished())
	saveResults(int(r), pln.GetResults())
}

func login() {
	for {
		login, password := getCredentials()
		phash := getPwdHash(password)
		if verifyCredentials(login, phash) {
			fmt.Println("Login successful")
			break
		} else {
			fmt.Println("Incorrect login or password, try again")
		}
	}
}

func getCredentials() (string, string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your login: ")
	scanner.Scan()
	login := scanner.Text()
	fmt.Print("Enter your password: ")
	scanner.Scan()
	password := scanner.Text()

	return login, password
}

func getPwdHash(password string) []byte {
	hash := sha256.New()

	pwdbytes := []byte(password)
	for i := 0; i < 100000; i++ {
		hash.Write(pwdbytes)
		pwdbytes = hash.Sum(nil)
	}

	return pwdbytes
}

func verifyCredentials(login string, phash []byte) bool {
	file, err := os.Open("credentials.bin")
	if err != nil {
		log.Fatal(err)
	}

	res, _ := ioutil.ReadAll(file)
	e := base64.StdEncoding.WithPadding(base64.NoPadding)

	creds := make([]byte, 512)
	n, _ := e.Decode(creds, res)
	creds = creds[:n]

	textCreds := strings.Split(string(creds), ":")

	return textCreds[0] == login && textCreds[1] == string(phash)
}

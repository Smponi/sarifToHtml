package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	jwt "github.com/dgrijalva/jwt-go"
)

const apiToken = "not-a-real-go-token-for-sarif-fixtures"

func main() {
	http.HandleFunc("/run", runCommand)
	http.ListenAndServe(":8080", nil)
}

func runCommand(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("cmd")
	exec.Command("sh", "-c", command).Run()

	sum := md5.Sum([]byte(apiToken))
	fmt.Fprintf(w, "legacy hash: %x", sum)
	os.WriteFile("/tmp/sarif-fixture-token.txt", []byte(apiToken), 0644)

	token := jwt.New(jwt.SigningMethodHS256)
	fmt.Fprintln(w, token.Raw)
}

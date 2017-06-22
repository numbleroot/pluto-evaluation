package worker

import (
	"bufio"
	"fmt"
	"log"

	"crypto/tls"

	"github.com/numbleroot/pluto-evaluation/imap-benchmark/config"
	"github.com/numbleroot/pluto-evaluation/imap-benchmark/sessions"
)

// Structs

// Session contains the user's credentials, an identifier for the
// session and a sequence of IMAP commands that has been generated
// by the sessions package.
type Session struct {
	User     string
	Password string
	ID       int
	Commands []sessions.IMAPCommand
}

// Functions

// Worker is the routine that sends the commands of the session
// to the server. The output will be logged and written in
// the logger channel.
func Worker(id int, config *config.Config, jobs chan Session, logger chan<- []string) {

	for job := range jobs {

		var output []string

		output = append(output, "########################")
		output = append(output, fmt.Sprintf("Session: %d", job.ID))
		output = append(output, fmt.Sprintf("User: %s", job.User))
		output = append(output, fmt.Sprintf("Password: %s", job.Password))
		output = append(output, "---- starting commands ----")

		// Connect to remote server.
		tlsConn, err := tls.Dial("tcp", config.Server.Addr, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Fatalf("Unable to connect to remote server %s: %v", config.Server.Addr, err)
		}

		conn := &Conn{
			c: tlsConn,
			r: bufio.NewReader(tlsConn),
		}

		// Login user for following IMAP commands session.
		conn.login(job.User, job.Password, id)

		for i := 0; i < len(job.Commands); i++ {

			switch job.Commands[i].Command {

			case "CREATE":

				command := fmt.Sprintf("%dX%d CREATE %dX%s", id, i, id, job.Commands[i].Arguments[0])

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("CREATE %d", respTime))

			case "DELETE":

				command := fmt.Sprintf("%dX%d DELETE %dX%s", id, i, id, job.Commands[i].Arguments[0])

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("DELETE %d", respTime))

			case "APPEND":

				command := fmt.Sprintf("%dX%d APPEND %dX%s %s %s", id, i, id, job.Commands[i].Arguments[0], job.Commands[i].Arguments[1], job.Commands[i].Arguments[2])

				respTime, err := conn.sendAppendCommand(command, job.Commands[i].Arguments[3])
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("APPEND %d", respTime))

			case "SELECT":

				command := fmt.Sprintf("%dX%d SELECT %dX%s", id, i, id, job.Commands[i].Arguments[0])

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("SELECT %d", respTime))

			case "STORE":

				command := fmt.Sprintf("%dX%d STORE %s FLAGS %s", id, i, job.Commands[i].Arguments[0], job.Commands[i].Arguments[1])

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("STORE %d", respTime))

			case "EXPUNGE":

				command := fmt.Sprintf("%dX%d EXPUNGE", id, i)

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("EXPUNGE %d", respTime))

			case "CLOSE":

				command := fmt.Sprintf("%dX%d CLOSE", id, i)

				respTime, err := conn.sendSimpleCommand(command)
				if err != nil {
					log.Fatal(err)
				}

				output = append(output, fmt.Sprintf("CLOSE %d", respTime))
			}
		}

		output = append(output, "########################")

		conn.logout(id)

		logger <- output
	}
}

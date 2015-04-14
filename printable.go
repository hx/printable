package main

import (
	"os"
	"os/signal"
	"regexp"
)

// Describes an upper limit of bytes that indicate a multi-byte UTF-8 sequence
type MultiByteLimits struct {
	ExtraLength  uint8
	MaximumValue byte
}

// Fill a buffer by reading from STDIN, panicking if it doesn't work
func read(b *[]byte) {
	_, err := os.Stdin.Read(*b)
	if err != nil {
		panic(err)
	}
}

func main() {

	// Go outputs "exist status 2" when you do a CTRL-C. We don't want that,
	// so we'll trap the interrupt signal and exit quietly when it's sent.
	sigint := make(chan os.Signal) // because we'll exit immediately, we don't need a buffer in this channel
	signal.Notify(sigint, os.Interrupt)

	// In the background, wait for the signal package to send something to
	// the sigint channel, and exit with code 0.
	go func() {
		<-sigint
		os.Exit(0)
	}()

	// Compile a list of multi-byte ranges for testing whether a byte starts
	// a UTF-8 sequence
	multiByteLimits := [5]MultiByteLimits{}
	for i := uint8(0); i < 5; i++ {
		multiByteLimits[i] = MultiByteLimits{
			ExtraLength:  i + 1,
			MaximumValue: ^byte(1 << (5 - i)),
		}
	}

	// The read buffer. We'll do one byte at a time.
	bytesRead := make([]byte, 1)

	// This regular expression tests if a byte can be printed.
	tester, _ := regexp.Compile("^[[:print:]\t\r\n]$")

	// Ad infinitum...
	for {

		// Read a byte from STDIN
		read(&bytesRead)

		// First, does this byte start a multi-byte sequence?
		if bytesRead[0] >= 0xC0 {

			// Find the length of the sequence based on the limit under which it falls
			for _, byteLimit := range multiByteLimits {

				if bytesRead[0] <= byteLimit.MaximumValue {

					// Read a few more bytes (based on the length for the detected range)
					extraBytes := make([]byte, byteLimit.ExtraLength)
					read(&extraBytes)

					// Write both the original byte and the extra bytes all at once (no need to make a new array)
					os.Stdout.Write(bytesRead)
					os.Stdout.Write(extraBytes)

					// Don't keep looking for more limits
					break
				}
			}
		} else {

			// Test for printability
			if tester.Match(bytesRead) {

				// It's printable. Print it.
				os.Stdout.Write(bytesRead)
			} else {

				// It's not printable. Print a cyan dot instead.
				os.Stdout.WriteString("\x1b[36m.\x1b[0m")
			}
		}
	}
}

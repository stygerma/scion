package main

import (
)


// Rule contains a rule for matching packets
type classRule struct {
	// This is currently means the ID of the sending border router
	sourceAs    	string
	nextHopAs		string
	destinationAs	string
	queueNumber 	int
}
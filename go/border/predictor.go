package main

import (
	"fmt"
	"math"
	//"errors"
)

type predictor struct {
	//intervalDur time.Duration //gives the length of the interval between BW measurements in Microseconds
	//predLen     time.Duration //
	firstPred  bool //
	secondPred bool
	//buffLen     int           //length of preds and trends arrays
	//preds       []int         //stores the predictions from now to (now + predLen)
	//trends      []int         //stores the trends from now to (now + predLen)
	newPred    int
	oldPred    int //pointer to the prediction furthest into the future
	newTrend   int //pointer to the trend furthest into the future
	oldTrend   int
	alpha      float64 //data smoothing factor
	beta       float64 //trend smoothing factor
	meanSquare int
}

func (p *predictor) DesUpdate(bw int) {
	if p.firstPred {
		fmt.Println("bw", bw)
		p.newTrend = bw
		p.firstPred = false
		fmt.Println("first", p.newTrend)
	} else if p.secondPred {
		fmt.Println("bw", bw)
		p.oldTrend = p.newTrend
		p.newPred = bw
		p.newTrend = bw - p.oldTrend
		p.secondPred = false
		fmt.Println("second", p.newTrend)
	} else {
		fmt.Println("bw", bw)
		p.oldPred = p.newPred
		p.oldTrend = p.newTrend
		p.newPred = int(p.alpha*float64(bw) + (1-p.alpha)*float64(p.oldPred+p.oldTrend))
		p.newTrend = int(p.beta*float64(p.newPred-p.oldPred) + (1-p.beta)*float64(p.oldTrend))
		p.meanSquare += int(math.Pow(float64(bw-p.newPred), 2))
		fmt.Printf("New smoothing value: %d, new trend value: %d, mean squared error: %d\n", p.newPred, p.newTrend, p.meanSquare)
	}
}

func (p predictor) DesPrediction(predLen int) int {
	return p.newPred + predLen*p.newTrend
}

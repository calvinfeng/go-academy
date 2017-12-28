package main

import "fmt"

type Player interface {
	GetMove(b *Board) (int, int, error)
	Mark() string
	Name() string
}

type HumanPlayer struct {
	name string
	mark string
}

func (p *HumanPlayer) GetMove(b *Board) (int, int, error) {
	fmt.Print("Enter position: ")
	var i, j int
	if n, err := fmt.Scanf("%d %d", &i, &j); err != nil || n != 2 {
		return 0, 0, err
	}

	fmt.Println("Your input:", i, j)
	return i, j, nil
}

func (p *HumanPlayer) Mark() string {
	return p.mark
}

func (p *HumanPlayer) Name() string {
	return p.name
}

// Bonus Phase
type ComputerPlayer struct {
	name string
	mark string
}

func (cp *ComputerPlayer) GetMove(b *Board) (int, int, error) {
	move := cp.minimax(b, cp.Mark(), 1)
	i, j := move["i"], move["j"]
	return i, j, nil
}

func (cp *ComputerPlayer) minimax(b *Board, mark string, depth int) map[string]int {
	if b.IsOver() {
		var score map[string]int
		score = make(map[string]int)
		if b.Winner() == cp.Mark() {
			score["value"] = 10 - depth
		} else {
			score["value"] = depth - 10
		}
		return score
	}

	scores := []map[string]int{}
	for _, pos := range b.GetAvailablePos() {
		newBoard := b.Copy()
		i, j := pos[0], pos[1]
		newBoard.PlaceMark(i, j, mark)

		var score map[string]int
		if mark == "X" {
			score = cp.minimax(newBoard, "O", depth+1)
		} else {
			score = cp.minimax(newBoard, "X", depth+1)
		}
		score["i"] = i
		score["j"] = j
		scores = append(scores, score)
	}

	if mark == cp.Mark() { // max
		maxScore := scores[0]
		for _, s := range scores {
			if maxScore["value"] < s["value"] {
				maxScore = s
			}
		}
		return maxScore
	} else { // min
		minScore := scores[0]
		for _, s := range scores {
			if minScore["value"] > s["value"] {
				minScore = s
			}
		}
		return minScore
	}
}

func (cp *ComputerPlayer) Mark() string {
	return cp.mark
}

func (cp *ComputerPlayer) Name() string {
	return cp.name
}

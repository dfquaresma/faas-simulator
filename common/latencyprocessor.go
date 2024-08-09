package common

import (
	"fmt"
	"math"
	"strconv"

	"github.com/emirpasic/gods/v2/trees/avltree"
)

type latencyProcessor struct {
	rp   *resourceProvisioner
	tree *avltree.Tree[string, float64]
}

func newLatencyProcessor(rp *resourceProvisioner) *latencyProcessor {
	return &latencyProcessor{
		rp:   rp,
		tree: avltree.New[string, float64](),
	}
}

func (lp *latencyProcessor) getCurrTLThreshould(i *invocation) (float64, error) {
	if lp.tree.Size() < 100 {
		return math.Inf(0), nil
	}

	dur := i.getDuration()
	if i.im.is_copy && lp.rp.cfg.Technique == "RequestHedgingOpt" {
		dur = dur - i.getTailLatencyThreshold()
	}
	lp.processLatency(dur)

	p, err := strconv.Atoi(lp.rp.cfg.TailLatencyProb[1:])
	if err != nil {
		return -1, fmt.Errorf("Error parsing tailLatencyProb", err)
	}

	return lp.getPercentileValue(p), nil
}

func (lp *latencyProcessor) processLatency(duration float64) {
	lp.tree.Put(fmt.Sprintf("%f_%d", duration, lp.tree.Size()), duration)
}

func (lp *latencyProcessor) getPercentileValue(p int) float64 {
	return kTh(lp.tree.Root, p, 0)
}

func kTh(n *avltree.Node[string, float64], k, prev int) float64 {
	curr := prev
	left := n.Children[0]
	if left != nil {
		curr = curr + left.Size()
		if curr >= k {
			return kTh(left, k, prev)
		}
	}

	right := n.Children[1]
	if right != nil && curr < k-1 {
		return kTh(right, k, curr+1)
	}

	return n.Value
}

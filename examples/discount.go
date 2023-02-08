package main

import (
	"fmt"
	"github.com/spikewong/gorule"
	"log"
	"os"
)

type User struct {
	balance     float64
	vipLevel    int
	inBlackList bool
}

func main() {
	// A simple rule engine used to calculate discount
	var (
		blacklistDiscount, vipDiscount int

		userInBlackList = &User{
			inBlackList: true,
		}
		vip = &User{
			vipLevel:    10,
			balance:     100,
			inBlackList: false,
		}

		blackListRule = gorule.NewRule(
			"black list rule",
			"inBlacklist",
			func(i interface{}) (interface{}, error) {
				return 0, nil
			},
		)
		vipRule = gorule.NewRule(
			"high level vip rule",
			"vipLevel >= 10 && !inBlacklist",
			func(i interface{}) (interface{}, error) {
				return 30, nil
			},
		)
	)

	discountEngine := gorule.NewEngine(
		gorule.WithLogger(log.New(os.Stdout, "discount", log.LstdFlags)),
		gorule.WithConfig(&gorule.Config{SkipBadRuleDuringMatch: false}),
	)
	for _, v := range []*gorule.Rule{vipRule, blackListRule} {
		if err := discountEngine.AddRule(v); err != nil {
			fmt.Printf("add rule failed: %v", err)
		}
	}

	matchedRules, err := discountEngine.Match(map[string]interface{}{
		"inBlacklist": userInBlackList.inBlackList,
		"balance":     userInBlackList.balance,
		"vipLevel":    userInBlackList.vipLevel,
	})
	if err != nil {
		fmt.Printf("encountered error when cal blacklist discount: %v \n", err)
	}
	for _, v := range matchedRules {
		d, _ := v.Execute(nil)
		discount, _ := d.(int)
		blacklistDiscount += discount
	}

	matchedRules, err = discountEngine.Match(map[string]interface{}{
		"inBlacklist": vip.inBlackList,
		"balance":     vip.balance,
		"vipLevel":    vip.vipLevel,
	})
	if err != nil {
		fmt.Printf("encountered error when cal vip discount: %v \n", err)
	}
	for _, v := range matchedRules {
		d, _ := v.Execute(nil)
		discount, _ := d.(int)
		vipDiscount += discount
	}

	fmt.Printf("blacklist discount: %v, vip discount: %v", blacklistDiscount, vipDiscount)
}

package main

import (
	"fmt"
	"sync"
)

type User struct {
	mu      sync.Mutex
	ID      string
	Name    string
	Balance float64
}

func (u *User) Deposit(amount float64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Balance += amount
}

func (u *User) Withdraw(amount float64) error {
	if u.Balance < amount {
		return fmt.Errorf("not enough balance")
	}
	u.mu.Lock()
	u.Balance -= amount
	u.mu.Unlock()
	return nil
}

type Transaction struct {
	ToID   string
	FromID string
	Amount float64
}

type PaymentSystem struct {
	Users        map[string]*User
	Transactions []Transaction
}

func (ps *PaymentSystem) AddUser(user *User) {
	ps.Users[user.ID] = user
}

func (ps *PaymentSystem) AddTransaction(transaction Transaction) {
	ps.Transactions = append(ps.Transactions, transaction)
}

func (ps *PaymentSystem) ProcessTransaction(transaction Transaction) error {
	fromUser, fromExists := ps.Users[transaction.FromID]
	toUser, toExists := ps.Users[transaction.ToID]
	if !fromExists {
		return fmt.Errorf("sending user not found")
	} else if !toExists {
		return fmt.Errorf("receiving user not found")
	} else if err := fromUser.Withdraw(transaction.Amount); err != nil {
		return err
	}
	toUser.Deposit(transaction.Amount)
	return nil
}

func main() {
	user1 := &User{ID: "1", Name: "Alice", Balance: 100.0}
	user2 := &User{ID: "2", Name: "Bob", Balance: 50.0}
	paymentSystem := &PaymentSystem{Users: make(map[string]*User)}
	paymentSystem.AddUser(user1)
	paymentSystem.AddUser(user2)
	fmt.Println("Before transaction:")
	fmt.Printf("Alice's balance: %.2f\n", user1.Balance)
	fmt.Printf("Bob's balance: %.2f\n", user2.Balance)
	transaction := Transaction{FromID: "1", ToID: "2", Amount: 30.0}
	paymentSystem.AddTransaction(transaction)
	/*
		if err := paymentSystem.ProcessTransaction(transaction); err != nil {
			fmt.Printf("Transaction failed: %s\n", err)
		} else {
			fmt.Println("After transaction:")
			fmt.Printf("Alice's balance: %.2f\n", user1.Balance)
			fmt.Printf("Bob's balance: %.2f\n", user2.Balance)
		}
	*/

	wrongTransaction := Transaction{FromID: "1", ToID: "2", Amount: 100.0}
	paymentSystem.AddTransaction(wrongTransaction)

	ch := make(chan Transaction, len(paymentSystem.Transactions))
	for _, transaction := range paymentSystem.Transactions {
		ch <- transaction
	}
	close(ch)

	wg := sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go paymentSystem.Worker(ch, &wg)
	}
	wg.Wait()
	/*
		for i, t := range paymentSystem.Transactions {
			fmt.Printf("Processing transaction %d\n", i+1)
			if err := paymentSystem.ProcessTransaction(t); err != nil {
				fmt.Printf("Transaction failed: %s\n", err)
			} else {
				fmt.Println("After transaction:")
				fmt.Printf("Alice's balance: %.2f\n", user1.Balance)
				fmt.Printf("Bob's balance: %.2f\n", user2.Balance)
			}
		}
	*/

}

func (ps *PaymentSystem) Worker(ch <-chan Transaction, wg *sync.WaitGroup) {
	for tr := range ch {
		fmt.Printf("Processing transaction from %s to %s for amount %.2f\n", tr.FromID, tr.ToID, tr.Amount)
		if err := ps.ProcessTransaction(tr); err != nil {
			fmt.Printf("Transaction failed: %s\n", err)
		} else {
			fmt.Println("After transaction:")
			for _, user := range ps.Users {
				fmt.Printf("%s's balance: %.2f\n", user.Name, user.Balance)
			}
		}
	}
	defer wg.Done()
}

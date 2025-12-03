package domain

import "time"

type User struct {
	ID        uint   
	Username  string     
	// Password  string       
	CreatedAt time.Time  
	UpdatedAt time.Time     
	Diaries   []Diary      
	Todos     []Todo    
	IsDeleted bool
	DeleteTime time.Time
}

package main 

import (
	"fmt" 
	"os"
)

func check(err error){
	if (err != nil){
		fmt.Println("Error Reading the file Contents")
	}
}


func main(){
	// first is path to the exe
	fileName := os.Args[1]

	content , err := os.ReadFile(fileName)
	byteC := len(content)
	check(err)
	line , word := wordCount(content)

	fmt.Print( line , word , byteC )
	fmt.Println(fileName)
}

func wordCount( arr []byte) (int , int ) {

	if ( len(arr) == 0 ) {
		return 0 , 0 
	}

	word_count := 0 
	new_line_count :=  0
	hasFound := false

	for _ , char := range string(arr) {

		if char == '\n' {
			new_line_count++
			hasFound = false
		} else if char == ' '  || char == '\t' || char == '\r' { 
			hasFound = false
		} else { 
			if (!hasFound){
				word_count++
	        	}
			hasFound = true 
		}
	}


	return new_line_count , word_count

}

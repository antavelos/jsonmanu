# JSON MANipUlator

**jsonmanu** is a Go library intended to be used a JSON data retrieval and update tool utilizing the
[JSONPath](https://goessner.net/articles/JsonPath/) notation.

## Installation
You can install the library using the go toolchain:
```shell
go get -u -v github.com/antavelos/jsonmanu
```

## Usage
The main api of the library is `jsonmanu.Get()` used to retrieve specific branches/leafs of JSON data as described by the provided JSONPath and `jsonmanu.Put()` used to update specific branches/leafs of JSON data as described by the provided JSONPath.

Following there are a few usage examples of the api. However, you can find the complete list of supported JSPNPath variations [below](#).

### Get
The signature of `Get` is `Get(data any, path string) (any, error)`.

#### Example
```go
import (
	"fmt"
	jm "github/antavelos/jsonmanu"
)

data := map[string]any{
	"store": map[string]any{
		"library": map[string]any{
			"books": []any{
				map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
				map[string]any{"author": "Schopenhauer", "title": "Book2", "price": 20},
				map[string]any{"author": "Stirner", "title": "Book3", "price": 15},
				map[string]any{"author": "Camus", "title": "Book4", "price": 5},
				map[string]any{"author": "Dostoevsky", "title": "Book5", "price": 10},
				map[string]any{"author": "Heraklitus", "title": "Book6", "price": 10},
			},
		},
	},
}

// get all authors
fmt.Println(jm.Get(data, "$.store..author"))
// ["Nietzsche" "Schopenhauer" "Stirner" "Camus" "Dostoevsky" "Heraklitus"]

// get the price of 1st and 6th books 
fmt.Println(jm.Get(data, "$..books[0,5].price"))
// [15 10]

// get the title of 2nd and 5th books
fmt.Println(jm.Get(data, "$..books[1:4].title"))
// ["Book2" "Book3" "Book4"]

// get the price of books authored by Nietzsche
fmt.Println(jm.Get(data, "$..books[?(@.author=Nietzsche)].price"))
// [15]
```
### Put
The signature of `Put` is `Put(data any, path string, value any) error`.

#### Example
```go
import (
	"fmt"
	jm "github/antavelos/jsonmanu"
)

data := map[string]any{
	"store": map[string]any{
		"library": map[string]any{
			"books": []any{
				map[string]any{"author": "Nietzsche", "title": "Book1", "price": 15},
				map[string]any{"author": "Schopenhauer", "title": "Book2", "price": 20},
				map[string]any{"author": "Stirner", "title": "Book3", "price": 15},
				map[string]any{"author": "Camus", "title": "Book4", "price": 5},
				map[string]any{"author": "Dostoevsky", "title": "Book5", "price": 10},
				map[string]any{"author": "Heraklitus", "title": "Book6", "price": 10},
			},
		},
	},
}
// update the author of the 2nd book to Nietzsche
jm.Put(data, "$..books[1].author", "Nietzsche")
fmt.Println(data)
// map[store:map[library:map[books:[
// map[author:Nietzsche price:15 title:Book1] 
// map[author:Nietzsche price:20 title:Book2] 
// map[author:Stirner price:15 title:Book3] 
// map[author:Camus price:5 title:Book4] 
// map[author:Dostoevsky price:10 title:Book5] 
// map[author:Heraklitus price:10 title:Book6]]]]]


// update the price of books authored by Camus to 30
jm.Put(data, "$..books[?(@.author=Camus)].price", 30)
fmt.Println(data)
// map[store:map[library:map[books:[
// map[author:Nietzsche price:15 title:Book1] 
// map[author:Nietzsche price:20 title:Book2] 
// map[author:Stirner price:15 title:Book3] 
// map[author:Camus price:30 title:Book4] 
// map[author:Dostoevsky price:10 title:Book5] 
// map[author:Heraklitus price:10 title:Book6]]]]]


```

## JSONPath usecases
Here is the complete list of the JSONPath supported or not yet usecases

| Expression | Description | Supported |
|------------|-------------|-----------|
| $ | The root object of array| YES |
| .property |	Selects the specified property in a parent object. | YES |
| ['property'] |	Selects the specified property in a parent object. | NO |
| [n] |	Selects the n-th element from an array. Indexes are 0-based. | YES |
| [index1,index2,...] |	Selects array elements with the specified indexes. Returns a list. | YES |
| ..property |	Recursive descent: Searches for the specified property name recursively and returns an array of all values with this property name. Always returns a list, even if just one property is found. | YES |
| * | Wildcard selects all elements in an object or an array, regardless of their names or indexes. For example, address.* means all properties of the address object, and book[\*] means all items of the book array. | YES |
| [start:end] [start:] | Selects array elements from the start index and up to, but not including, end index. If end is omitted, selects all elements from start until the end of the array. Returns a list. | YES |
| [:n] |	Selects the first n elements of the array. Returns a list. | YES |
| [-n:] |	Selects the last n elements of the array. Returns a list. | NO |
| [?(expression)] |	Filter expression. Selects all elements in an object or array that match the specified filter. Returns a list.| YES |
| @	| Used in filter expressions to refer to the current node being processed. | YES |

## Testing
```shell
go test -v
```

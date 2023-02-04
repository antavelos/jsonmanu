# JSON MANipUlator v1.1.1
[![Actions Status](https://github.com/antavelos/jsonmanu/workflows/Testing/badge.svg)](https://github.com/antavelos/jsonmanu/actions)
[![GoDoc](https://godoc.org/github.com/antavelos/jsonmanu?status.svg)](https://godoc.org/github.com/antavelos/jsonmanu)

**jsonmanu** is a Go library intended to be used as a JSON data retrieval and update tool utilizing the
[JSONPath](https://goessner.net/articles/JsonPath/) notation.

#### Table of contents
- [JSON MANipUlator v1.1.0](#json-manipulator-v110)
			- [Table of contents](#table-of-contents)
	- [Installation](#installation)
	- [JSONPath usecases](#jsonpath-usecases)
	- [Filtering with expressions](#filtering-with-expressions)
	- [Usage](#usage)
		- [Get](#get)
		- [Put](#put)
	- [Testing](#testing)


## Installation
You can install the library using the go toolchain:
```shell
go get -u -v github.com/antavelos/jsonmanu
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
| [?(expression)] |	Filter expression. Selects all elements in an object or array that match the specified filter. Returns a list. You can see mre details in the [below](#filtering-with-expressions) section.| YES |
| @	| Used in filter expressions to refer to the current node being processed. | YES |

## Filtering with expressions
With an expression you can filter array elements bases of the properties of its object items. The supported operators are `==`, `!=`, `<`, `>`, `<=`, `>=` and they apply on both numbers an strings. 

Examples:
* `$.books[?(author == "Nietzsche")]`  filters all the book authored by Nietzsche. 
* `$.books[?(author != "Nietzsche")]` filters all the books except from those authored by Nietzsche.
* `$.books[?(price == 10)]` filters all the books with price equal to 10.
* `$.books[?(price != 10)]` filters all the books with price not equal to 10.
* `$.books[?(price >= 10)]` filters all the books with price greater or equal to 10.
* `$.books[?(price <= 10)]` filters all the books with price less or equal to 10.
* `$.books[?(price > 10)]` filters all the books with price greater than 10.
* `$.books[?(price < 10)]` filters all the books with price less than 10.

## Usage
The main api of the library is `jsonmanu.Get()` used to retrieve specific branches/leafs of JSON data as described by the provided JSONPath and `jsonmanu.Put()` used to update specific branches/leafs of JSON data as described by the provided JSONPath.

Following there are a few usage examples of the api. However, you can find the complete list of supported JSPNPath variations [below](#jsonpath-usecases).

### Get
The signature of `Get` is `Get(data any, path string) (any, error)`.

It accepts:
* `data` of type `any` as it can be any unstructured data unmarshaled in an `interface{}`
* `path` which must be a string complying with the [JSONPath usecases](#jsonpath-usecases).

It returns:
* a value of type `any` as it can be any of the types handled by [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal) function.
* an error in case something went wrong either during the path parsing or the value retrieval.

Here is an example:
```go
package main

import (
	"encoding/json"
	"fmt"

	jm "github.com/antavelos/jsonmanu"
)

var jsonStr = []byte(`{
	"store": {
		"library": {
			"books": [
				{
					"author": "Nietzsche", 
					"title": "Book1", 
					"price": 15
				},
				{
					"author": "Schopenhauer", 
					"title": "Book2", 
					"price": 20
				},
				{
					"author": "Stirner", 
					"title": "Book3", 
					"price": 15
				},
				{
					"author": "Camus", 
					"title": "Book4", 
					"price": 5
				},
				{
					"author": "Dostoevsky", 
					"title": "Book5", 
					"price": 10
				},
				{
					"author": "Heraklitus", 
					"title": "Book6", 
					"price": 10
				}
			]
		}
	}
}`)

func main() {


	var data interface{}

	err := json.Unmarshal(jsonStr, &data)
	if err != nil {
		panic(err)
	}

	// get all authors
	fmt.Println(jm.Get(data, "$.store..author"))
	// ["Nietzsche" "Schopenhauer" "Stirner" "Camus" "Dostoevsky" "Heraklitus"] <nil>

	// get the price of 1st and 6th books
	fmt.Println(jm.Get(data, "$..books[0,5].price"))
	// [15 10] <nil>

	// get the title of 2nd and 5th books
	fmt.Println(jm.Get(data, "$..books[1:4].title"))
	// ["Book2" "Book3" "Book4"] <nil>

	// get the price of books authored by Nietzsche
	fmt.Println(jm.Get(data, "$..books[?(@.author == Nietzsche)].price"))
	// [15] <nil>
}
```
### Put
The signature of `Put` is `Put(data any, path string, value any) error`.

It accepts:
* `data` of type `any` as it can be any unstructured data unmarshaled in an `interface{}`
* `path` of type `string` complying with the [JSONPath usecases](#jsonpath-usecases).
* `value` of type `any` as it can be any of the types handled by [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal) function.

It returns:
* an error in case something went wrong either during the path parsing or the data update.

After the execution the `data` variable will contain the update version of it.

Here is an example
```go
package main

import (
	"encoding/json"
	"fmt"

	jm "github.com/antavelos/jsonmanu"
)

var jsonStr = []byte(`{
	"store": {
		"library": {
			"books": [
				{
					"author": "Nietzsche", 
					"title": "Book1", 
					"price": 15
				},
				{
					"author": "Schopenhauer", 
					"title": "Book2", 
					"price": 20
				},
				{
					"author": "Stirner", 
					"title": "Book3", 
					"price": 15
				},
				{
					"author": "Camus", 
					"title": "Book4", 
					"price": 5
				},
				{
					"author": "Dostoevsky", 
					"title": "Book5", 
					"price": 10
				},
				{
					"author": "Heraklitus", 
					"title": "Book6", 
					"price": 10
				}
			]
		}
	}
}`)

func main() {


	var data interface{}

	err := json.Unmarshal(jsonStr, &data)
	if err != nil {
		panic(err)
	}

	// update the author of the 2nd book to Nietzsche
	jm.Put(data, "$..books[1].author", "Nietzsche")
	fmt.Println(data)
	// map[store:map[library:map[books:[
	// map[author:Nietzsche price:15 title:Book1] 
	// map[author:Nietzsche price:20 title:Book2] <---------
	// map[author:Stirner price:15 title:Book3] 
	// map[author:Camus price:5 title:Book4] 
	// map[author:Dostoevsky price:10 title:Book5] 
	// map[author:Heraklitus price:10 title:Book6]]]]]

	// update the price of books authored by Camus to 30
	jm.Put(data, "$..books[?(@.author == Camus)].price", 30)
	fmt.Println(data)
	// map[store:map[library:map[books:[
	// map[author:Nietzsche price:15 title:Book1] 
	// map[author:Nietzsche price:20 title:Book2] 
	// map[author:Stirner price:15 title:Book3] 
	// map[author:Camus price:30 title:Book4] <---------
	// map[author:Dostoevsky price:10 title:Book5] 
	// map[author:Heraklitus price:10 title:Book6]]]]]
}

```
  
## Testing
```shell
go test -v
```

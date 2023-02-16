# JSON MANipUlator (jsonmanu)
[![Actions Status](https://github.com/antavelos/jsonmanu/workflows/Testing/badge.svg)](https://github.com/antavelos/jsonmanu/actions)
[![GoDoc](https://godoc.org/github.com/antavelos/jsonmanu?status.svg)](https://godoc.org/github.com/antavelos/jsonmanu)

**jsonmanu** is a Go library intended to be used as a JSON data manipulator by retrieving and updating JSON object sub-branches as well as mapping data between two JSON objects. The data querying is done by utilizing the 
[JSONPath](https://goessner.net/articles/JsonPath/) notation.

#### Table of contents
- [JSON MANipUlator](#json-manipulator)
			- [Table of contents](#table-of-contents)
	- [Installation](#installation)
	- [Testing](#testing)
	- [API](#api)
		- [`Get(data map[string]any, path string) (any, error)`](#getdata-mapstringany-path-string-any-error)
		- [`Put(data map[string]any, path string, value any) error`](#putdata-mapstringany-path-string-value-any-error)
		- [`Map(src map[string]any, dst map[string]any, mappers []Mapper) []error`](#mapsrc-mapstringany-dst-mapstringany-mappers-mapper-error)
		- [Transformation](#transformation)
		- [Transformer](#transformer)
			- [`SplitTransformer`](#splittransformer)
			- [`JoinTransformer`](#jointransformer)
			- [`ReplaceTransformer`](#replacetransformer)
			- [`StringMatchTransformer`](#stringmatchtransformer)
			- [`SubStrTransformer`](#substrtransformer)
			- [`NumberTransformer`](#numbertransformer)
	- [JSONPath usecases](#jsonpath-usecases)
		- [Filtering with expressions](#filtering-with-expressions)
	- [LICENSE](#license)

## Installation
You can install the library using the go toolchain:
```shell
go get -u -v github.com/antavelos/jsonmanu
```

## Testing
```shell
go test -v
```

## API
The main api of the library consists of:
* [Get](#get) which is used to retrieve specific branches/leafs of JSON data as described by the provided JSONPath
* [Put](#put) which is used to update specific branches/leafs of JSON data as described by the provided JSONPath.
* [Map](#map) which is used to map data from a JSON data source to a JSON data destination. 

In the following sections describe the library api in details along with examples. For more information about the supported JSPNPath variations please refer [below](#jsonpath-usecases).

### `Get(data map[string]any, path string) (any, error)`

It accepts:
* `data` of type `map[string]any` the typical type of an unmarshalled JSON object. 
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


	var data any

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
### `Put(data map[string]any, path string, value any) error`
It accepts:
* `data` of type `map[string]any` the typical type of an unmarshalled JSON object. 
* `path` of type `string` complying with the [JSONPath usecases](#jsonpath-usecases).
* `value` of type `any` as it can be any of the types handled by [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal) function.

It returns:
* an error in case something went wrong either during the path parsing or the data update.

After the call the `data` variable will contain the update version of it.

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


	var data any

	err := json.Unmarshal(jsonStr, &data)
	if err != nil {
		panic(err)
	}

	// update the author of the 2nd book to Nietzsche
	jm.Put(data, "$..books[1].author", "Nietzsche")

	fmt.Println(data["store"]["library"]["books"][1])
	// map[author:Nietzsche price:20 title:Book2]

	// update the price of books authored by Camus to 30
	jm.Put(data, "$..books[?(@.author == Camus)].price", 30)

	fmt.Println(data["store"]["library"]["books"][3])
	// map[author:Camus price:30 title:Book4]
}

```

### `Map(src map[string]any, dst map[string]any, mappers []Mapper) []error`
It accepts:
* `src` of type `map[string]any` which is the source object to be mapped. 
* `dst` of type `map[string]any` which is the destination object where source will be mapped to. It cannot be nil.
* `mappers` of type `Mapper` which is a list of JSONPath pairs (source and destination) along with one or more optional [transformations](#transformations) that will apply (in chain mode and in the order of cofiguration) on the retrieved source value before its put in the destionation object.
  
It returns:
* a list of errors per mapper. 

After the call, the changes in `dst` apply in place.

Here is an example:
```go
package main

import (
	"encoding/json"
	"fmt"

	jm "github.com/antavelos/jsonmanu"
)

var sourceJsonStr = []byte(`{
	"store": {
		"library": {
			"books": [
				{
					"author": "Nietzsche", 
					"title": "Book1", 
					"price": 15,
					"summary": "born on 15/10/1844"
				},
				{
					"author": "Schopenhauer", 
					"title": "Book2", 
					"price": 20,
					"summary": "born on 22/02/1788"
				},
				{
					"author": "Stirner", 
					"title": "Book3", 
					"price": 15,
					"summary": "born on 25/10/1806"
				},
				{
					"author": "Camus", 
					"title": "Book4", 
					"price": 5,
					"summary": "born on 07/11/1913"
				},
				{
					"author": "Dostoevsky", 
					"title": "Book5", 
					"price": 10,
					"summary": "born on 30/10/1821"
				}
			]
		}
	}
}`)

func main() {
	var sourceData map[string]any
	destData := make(map[string]any)

	err := json.Unmarshal(sourceJsonStr, &sourceData)
	if err != nil {
		panic(err)
	}

	mappers := []jm.Mapper{
		jm.Mapper{
			SrcJsonPath: "$.store.library.books.author",
			DstJsonPath: "$.info.authors",
		},
		jm.Mapper{
			SrcJsonPath: "$.store.library.books.summary",
			DstJsonPath: "$.info.birthYears",
			Transformations: []jm.Transformation{
				{Trsnfmr: jm.StringMatchTransformer{Regex: `\d{2}/\d{2}/\d{4}`}},
				{Trsnfmr: jm.SplitTransformer{Delim: "/", Index: 2}},
				{Trsnfmr: jm.NumberTransformer{}},
			},
		},
	}

	jm.Map(sourceData, destData, mappers)

	prettyDstData, _ := json.MarshalIndent(destData, "", "  ")
	fmt.Printf("%s", prettyDstData)
	// {
	// 	"info": {
	// 	  "authors": [
	// 		"Nietzsche",
	// 		"Schopenhauer",
	// 		"Stirner",
	// 		"Camus",
	// 		"Dostoevsky"
	// 	  ],
	// 	  "birthYears": [
	// 		1844,
	// 		1788,
	// 		1806,
	// 		1913,
	// 		1821
	// 	  ]
	// 	}
	// }
}


```

### Transformation
Transformation is a struct type that contains:
* a [Transformer](#transformer) and 
* an `asArray` flag: Since a retrieved source value can often be an array, this flag can be used to indicate whether the transformation will apply on the source value as an array or it will apply on each item individualy.

### Transformer
Transformer is a interface that implements the `Transform(value any) (any, error)` method. That means that it's up to the user to define their custom transformers in order to manipulate the retrieved data. 
However, there are already some pre-built transformers that can be mainly used for string manipulation purposes:

#### `SplitTransformer`
```go
type SplitTransformer struct {
	Delim string
	Index int
}
```
`SplitTransformer` will split a string value based on the provided delimeter and it will pick the element defined by the provided index from the occured array.

#### `JoinTransformer`
```go
type JoinTransformer struct {
	Delim string
}
```
`JoinTransformer` joins the values of an array based on the provided delimiter.

#### `ReplaceTransformer`
```go
type ReplaceTransformer struct {
	OldVal string
	NewVal string
}
```

`ReplaceTransformer` replaces a substring in a string value with another.

#### `StringMatchTransformer`
```go
type StringMatchTransformer struct {
	Regex string
}
```
`StringMatchTransformer` returns a regex matched substring of a string.

#### `SubStrTransformer`
```go
type SubStrTransformer struct {
	Start int
	End int
}
```
`SubStrTransformer` returns a slice of a string value based on the provided indices.

#### `NumberTransformer`
```go
type NumberTransformer struct{}
```
`NumberTransformer` converts a string value to float64.

## JSONPath usecases
Here is the complete list of the JSONPath supported (or not yet) usecases:

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

### Filtering with expressions
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

## LICENSE
See LICENSE file.
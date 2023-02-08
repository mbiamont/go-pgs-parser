# go-pgs-parser

Golang PGS Parser

## Installation

```shell
go get github.com/mbiamont/go-pgs-parser
```

## Getting started

The parser requires you specify a SUP file's path.

More info about PGS and SUP file here: https://fileinfo.com/extension/sup

For each subtitle bitmap, it'll ask you in which file to write it.

### Save as PNG images (recommended)

```go
package main

import (
	"fmt"
	"github.com/mbiamont/go-pgs-parser/pgs"
	"os"
)

func main() {
	parser := pgs.NewPgsParser()

	parser.ConvertToPngImages("./sample/input.sup", func(index int) (*os.File, error) {
		return os.Create(fmt.Sprintf("./sample/subs/input.%d.png", index))
	})
}
```
### Save as JPG images

```go
package main

import (
	"fmt"
	"github.com/mbiamont/go-pgs-parser/pgs"
	"os"
)

func main() {
	parser := pgs.NewPgsParser()

	parser.ConvertToJpgImages("./sample/input.sup", func(index int) (*os.File, error) {
		return os.Create(fmt.Sprintf("./sample/subs/input.%d.jpg", index))
	})
}
```

### Output example

<img src="./art/output-example.png" />


## Extract SUP from MKV

You can extract a SUP file using ffmpeg like this:

```shell
ffmpeg -i input.mkv -map 0:s:0 -c copy input.sup -y
```

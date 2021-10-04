# filmdetect

Filmdetect is a cli tool, and a library for detecting what film recipe was used
to create a Fujifilm jpeg file.

## recipes

You will need a directory of recipe files.  You can create your own, or [use one I maintain][1].

## cli

```
$ filmdetect --simulation-dir "path/to/simulation/dir" <some fujifilm jpeg file>
Kodak Portra 400
```

## library

```go
package main

import (
    "os"
    "github.com/honza/filmdetect"
)

func main() {
    file, err := os.Open("some-fujifilm-file.jpg")

    if err != nil {
        return
    }

    simulationDir := "path/to/simulations"
    diffs, havePerfectMatch, err := filmdetect.Detect(simulationDir, file)

    if err != nil {
        return
    }
    
    if havePerfectMatch {
        fmt.Println(diffs[0].Candidate.Name)
    } else {
        for _, diff := range diffs {
            fmt.Println(diff)
        }
    }
    
    
}
```

## license

GPLv3

[1]: https://github.com/honza/film-simulations

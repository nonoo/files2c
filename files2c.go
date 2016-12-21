package main

import(
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"log"
	"flag"
	"regexp"
	"strings"
	"strconv"
)

var main_colCount = 15
var main_dir string
var main_outHeaderFilename = "out.h"
var main_outHeaderFile *os.File
var main_outModuleFilename = "out.c"
var main_outModuleFile *os.File

func process(fi os.FileInfo) {
	fmt.Println("  processing " + fi.Name())

	f, err := os.Open(main_dir + "/" + fi.Name())
	if err != nil {
		fmt.Println("    error opening file, skipping")
		return
	}
	defer f.Close()

	// Replacing dots to underscores in module name.
	moduleVarName := strings.ToLower(strings.Replace(fi.Name(), ".", "_", -1))
	if _, err := strconv.Atoi(string([]rune(moduleVarName)[0])); err == nil {
		// If the module name starts with a number, we prepend an underscore.
		moduleVarName = "_" + moduleVarName
	}

	arrayDefine := "uint8_t " + moduleVarName + "[" + strconv.FormatInt(fi.Size(), 10) + "]"

	out := "\n" + arrayDefine + " = {\n\t"
	main_outModuleFile.WriteString(out)

	b := make([]byte, 1) // We read to this 1 byte buffer.
	i := 0
	for ; true; i++ {
		bytesRead, err := f.Read(b)
		if bytesRead != 1 {
			if err != io.EOF {
				log.Panic("    error during file read")
				return
			}
			break
		}

		if i > 0 {
			main_outModuleFile.WriteString(", ")

			if i % main_colCount == 0 {
				main_outModuleFile.WriteString("\n\t")
			}
		}

		out := fmt.Sprintf("0x%.2x", b)
		main_outModuleFile.WriteString(out)
	}
	// If we have finished by NOT adding a new line.
	if i % main_colCount != 0 {
		main_outModuleFile.WriteString("\n")
	}
	main_outModuleFile.WriteString("};\n")

	main_outHeaderFile.WriteString("extern " + arrayDefine + ";\n")
}

func initHeaderFile() {
	main_outHeaderFile.WriteString("#ifndef " + strings.ToUpper(strings.Replace(main_outHeaderFilename, ".", "_", -1)) + "__\n")
	main_outHeaderFile.WriteString("#define " + strings.ToUpper(strings.Replace(main_outHeaderFilename, ".", "_", -1)) + "__\n\n")
	main_outHeaderFile.WriteString("#include <stdint.h>\n\n")
}

func initModuleFile() {
	main_outModuleFile.WriteString("#include \"" + main_outHeaderFilename + "\"\n");
}

func main() {
	flag.IntVar(&main_colCount, "c", main_colCount, "hex value column count")
	flag.StringVar(&main_dir, "d", main_dir, "convert files in this directory")
	flag.StringVar(&main_outHeaderFilename, "h", main_outHeaderFilename, "output header filename")
	flag.StringVar(&main_outModuleFilename, "m", main_outModuleFilename, "output module filename")
	flag.Parse()

	if main_dir == "" {
		log.Fatal("no convert directory defined, see help")
	}

	var err error
	main_outHeaderFile, err = os.Create(main_outHeaderFilename)
	if err != nil {
		log.Fatal("can't create header file " + main_outHeaderFilename)
	}
	defer main_outHeaderFile.Close()
	initHeaderFile()

	main_outModuleFile, err = os.Create(main_outModuleFilename)
	if err != nil {
		log.Fatal("can't create module file " + main_outModuleFilename)
	}
	defer main_outModuleFile.Close()
	initModuleFile()

	files, err := ioutil.ReadDir(main_dir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("files2c processing directory " + main_dir)

	// Ignoring .go files.
	regexpGo, _ := regexp.Compile("\\.go$")

    for _, fi := range files {
    	if !fi.Mode().IsRegular() {
    		continue
    	}
    	if fi.Name() == main_outHeaderFilename || fi.Name() == main_outModuleFilename {
    		continue
    	}
    	if regexpGo.MatchString(fi.Name()) {
    		continue
    	}
		process(fi)
	}

	main_outHeaderFile.WriteString("\n#endif\n")

	fmt.Println("files2c done")
}

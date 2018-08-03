package generate

import (
    "bufio"
    "flag"
    "fmt"
    "io/ioutil"
    "math/rand"
    "os"
    "time"
)

// Colors and other printing items
const CLR_RED =   "\x1b[31;1m"
const CLR_GRN =   "\x1b[32;1m"
const CLR_END =   "\x1b[0m"
const ERR_ICON =  "[!]"
const INFO_ICON = "[+]"

func Generate(domain string, category string, template string) {
    delErr := os.Remove(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/CNAME")
    checkErr(delErr)
    f, crtErr := os.Create(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/CNAME")
    checkErr(crtErr)
    fmt.Fprintln(f, domain)

    // Initialize global pseudo random generator
    rand.Seed(time.Now().UnixNano())

    cat := "finance"
    if category != "" {
        cat = category
    }

    var layout string
    if cat == "finance" {
        layout = generateIndex(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/jekyll-templates/finance", template)
    } else if cat == "healthcare" {
        layout = generateIndex(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/jekyll-templates/healthcare", template)
    } else {
        errStr := "The category you entered is invalid."
        flag.PrintDefaults()
        fmt.Printf("%s%s%s %s\n", CLR_RED, ERR_ICON, CLR_END, errStr)
        os.Exit(1)
    }

    setThemeColor(layout)
}

// generateIndex creates the index.md file to be
// used as index.html
func generateIndex(path string, templatePath string) (lo string) {
        var lines []string
        var layout string
        if templatePath == "" {
            layout = randFromFile(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/bslayouts")
            imgOne := randFile(path + "/img")
            imgOneStr := "imgOne: ." + imgOne.Name()
            imgTwo := randFile(path + "/img")
            imgTwoStr := "imgTwo: ." + imgTwo.Name()
            imgThree := randFile(path + "/img")
            imgThreeStr := "imgThree: ." + imgThree.Name()
            imgFour := randFile(path + "/img")
            imgFourStr := "imgFour: ." + imgFour.Name()
            imgsStr := imgOneStr + "\n" + imgTwoStr + "\n" + imgThreeStr + "\n" + imgFourStr

            lines = append(lines, "---")
            lines = append(lines, "layout: " + layout)
            lines = append(lines, imgsStr)
            lines = append(lines, "title: " + randFromFile(path + "/titles"))
            title := randFromFile(path + "/titles")
            lines = append(lines, "navTitle: " + title)
            lines = append(lines, "heading: " + title)
            lines = append(lines, "subheading: " + randFromFile(path + "/subheading"))
            lines = append(lines, "aboutHeading: About Us")
            lines = append(lines, generateServices(path + "/services"))
            lines = append(lines, generateCategories(path + "/categories"))
            lines = append(lines, "servicesHeading: Our offerings")
            lines = append(lines, "contactDesc: Contact Us Today!")
            lines = append(lines, "phoneNumber: " + randFromFile(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/phone-num"))
            lines = append(lines, "email: " + randFromFile(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/emails"))
            lines = append(lines, "---")
            lines = append(lines, "\n")
            lines = append(lines, randFromFile(path + "/content"))
        } else {
            template, err := os.Open(templatePath)
            checkErr(err)
            scanner := bufio.NewScanner(template)
            for scanner.Scan() {
                lines = append(lines, scanner.Text())
            }
        }

        writeTemplate(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/index.md", lines)

        return layout
}

// randFromFile returns a random line from
// a specified file
func randFromFile(path string) (line string) {
    file, err := os.Open(path)
    checkErr(err)

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    return lines[rand.Intn(len(lines))]
}

// generateServices randomly selects service pairs
// for the index.md file
func generateServices(path string) (line string) {
    file, err := os.Open(path)
    checkErr(err)

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    selInts := rand.Perm(len(lines)/2)
    selInts = selInts[0:4]

    return "serviceOne: " + lines[selInts[0]*2] + "\n" +
           "serviceOneDesc: " + lines[(selInts[0]*2)+1] + "\n" +
           "serviceTwo: " + lines[selInts[1]*2] + "\n" +
           "serviceTwoDesc: " + lines[(selInts[1]*2)+1] + "\n" +
           "serviceThree: " + lines[selInts[2]*2] + "\n" +
           "serviceThreeDesc: " + lines[(selInts[2]*2)+1] + "\n" +
           "serviceFour: " + lines[selInts[3]*2] + "\n" +
           "serviceFourDesc: " + lines[(selInts[3]*2)+1]
}

// generateCategories randomly selects category
// pairs for the index.md file
func generateCategories(path string) (line string) {
    file, err := os.Open(path)
    checkErr(err)

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    selInts := rand.Perm(len(lines)/2)
    selInts = selInts[0:4]

    return "categoryOne: " + lines[selInts[0]*2] + "\n" +
           "categoryOneName: " + lines[(selInts[0]*2)+1] + "\n" +
           "categoryTwo: " + lines[selInts[1]*2] + "\n" +
           "categoryTwoName: " + lines[(selInts[1]*2)+1] + "\n" +
           "categoryThree: " + lines[selInts[2]*2] + "\n" +
           "categoryThreeName: " + lines[(selInts[2]*2)+1] + "\n" +
           "categoryFour: " + lines[selInts[3]*2] + "\n" +
           "categoryFourName: " + lines[(selInts[3]*2)+1]
}

// randFile randomly selects and returns a file from the directory
// that the path provided points to.  If the randomly selected file
// is a directory, it will pick again until it is a regular file.
func randFile(path string) (file *os.File) {
    files, err := ioutil.ReadDir(path)
    checkErr(err)

    // Randomly select a file
    var p string
    isDir := true
    for isDir {
        p = path + "/" + files[rand.Intn(len(files))].Name()
        fi, err := os.Stat(p)
        checkErr(err)
        switch mode := fi.Mode(); {
            case mode.IsDir():
            case mode.IsRegular():
                isDir = false
        }
    }
    file, er := os.Open(p)
    checkErr(er)
    return file
}

// writeTemplate writes the template file to the main Jekyll
// directory for use with index.html
func writeTemplate(path string, lines []string) {
    delErr := os.Remove(path)
    checkErr(delErr)
    f, crtErr := os.Create(path)
    checkErr(crtErr)

    for _, line := range lines {
        fmt.Fprintln(f, line)
    }
}

// setThemeColor selects a random color to
// be used in the theme of the website
func setThemeColor(layout string) {
    var path string
    switch layout {
        case "index":
            path = homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/bootstrap/css/creative.css"
            break
        case "index2":
            path = homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/bootstrap2/css/resume.css"
            break
        case "index3":
            path = homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/bootstrap3/css/stylish-portfolio.css"
            break
        case "":
            return
            break
        default:
            fmt.Fprintf(os.Stderr, "%s%s Sphynx%s: %s\n", CLR_RED, ERR_ICON, CLR_END, "Layout name not found")
            os.Exit(1)
    }

    file, err := os.Open(path)
    checkErr(err)

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    lines = lines[0:(len(lines)-3)]
    clr := randFromFile(homeDir + "/go/src/git.praetorianlabs.com/mars/sphinx/colors")
    lines = append(lines, ":root {\n\t--themeColor: " + clr + ";\n}")
    writeTemplate(path, lines)
}

// checkErr is an abstracted generic error checking func.
func checkErr(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s%s Sphynx%s: %v\n", CLR_RED, ERR_ICON, CLR_END, err)
        os.Exit(1)
    }
}

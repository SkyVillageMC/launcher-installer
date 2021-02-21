package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

var (
	data     *string
	a        fyne.App
	progress float64
	pBar     *widget.ProgressBar
	pSize    uint64
	pOffset  float32
)

type WriteCounter struct {
	Total uint64
	Max   uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	progress = float64(wc.Total) / float64(wc.Max)
	(*pBar).SetValue((progress / 10 * float64(pSize)) + float64(pOffset))
}

func main() {
	log.Println("Launching SkyLauncher")
	appdata := os.Getenv("APPDATA")

	data = &appdata

	if exists(appdata + "\\.skyvillage\\launcher.jar") {
		launch()
	} else {
		launchInstaller()
	}
}

func startInstaller() {
	//Check if .skyvillage exists
	var fileInfo, err = os.Stat(*data + "\\.skyvillage")
	if err != nil || !fileInfo.IsDir() {
		error := os.Mkdir(*data+"\\.skyvillage", os.ModeDir)
		if error != nil {
			log.Fatalln("Can not create directory .skyvillage")
			os.Exit(500)
		}
	}
	log.Println("Found .skyvillage!")
	checkJava()
}

func launch() {
	proc := exec.Command(*data+"\\.skyvillage\\java\\jre\\bin\\java.exe", "-cp", *data+"\\.skyvillage\\launcher.jar", "hu.bendi.skylauncher.Launcher")
	env := proc.Env
	proc.Dir = *data + "\\.skyvillage"
	env = append(proc.Env, "JAVA_HOME="+*data+"\\.skyvillage\\java")
	env = append(env, "APPDATA="+*data)
	proc.Env = env
	proc.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT,
	}

	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	err := proc.Run()

	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}

func unzip(from string, targetDir string) {
	zipReader, err := zip.OpenReader(from)
	if err != nil {
		log.Fatal(err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.Reader.File {
		zippedFile, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer zippedFile.Close()

		extractedFilePath := filepath.Join(
			targetDir,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			log.Println("Creating directory:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("Extracting file:", file.Name)

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				log.Println(err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func downloadLauncher() {
	url := "https://bendimester23.tk/assets/launcher.jar"
	downloadFile(*data+"\\.skyvillage\\launcher.jar", url, 2, 0.7)
}

func downloadJava() {
	log.Println("Downloading java...")
	url := "https://www.dropbox.com/s/j9wniroiggs5vkn/java.zip?dl=1"
	erro := mdIfNotPresent(*data + "\\.skyvillage\\tmp")
	if erro != nil {
		log.Println("Error creating directory")
		log.Println("Error: ", erro)
		os.Exit(500)
	}
	err := downloadFile(*data+"\\.skyvillage\\tmp\\java.zip", url, 7, 0)
	if err != nil {
		log.Printf("Error: %s", err)
		log.Println("Error while downloading java. Please contact support.")
	}
	log.Println("Installing java...")
	unzip(*data+"\\.skyvillage\\tmp\\java.zip", *data+"\\.skyvillage\\")
	rErro := os.Remove(*data + "\\.skyvillage\\tmp\\java.zip")
	if rErro != nil {
		log.Println("Error deleting file")
	}
	checkJava()
}

func exists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return true
}

func mdIfNotPresent(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		os.MkdirAll(filepath, os.ModeDir)
	}
	return err
}

func downloadFile(filepath string, url string, size uint64, offset float32) error {
	pSize = size
	pOffset = offset
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{}
	counter.Max = uint64(resp.ContentLength)

	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}

func checkJava() bool {
	data := os.Getenv("APPDATA")
	_, err := os.Stat(data + "\\.skyvillage\\java\\bin\\java.exe")
	if err != nil {
		log.Println("Java not found, downloading...")
		return false
	}

	proc := exec.Command(data+"\\.skyvillage\\java\\bin\\java.exe", "--version")
	erro := proc.Start()
	if erro != nil {
		log.Println("Java corrupted, downloading...")
		return false
	}
	return true
}

func launchInstaller() {
	a = app.NewWithID("hu.bendi.skylauncher")

	log.Println("Starting installer...")
	drv := a.Driver()
	w := drv.CreateWindow("SkyVillage Telepítő")

	w.CenterOnScreen()
	w.SetFixedSize(true)
	label := widget.NewLabelWithStyle(" \nÜdvözöllek!\n \n \n ", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	w.SetContent(widget.NewVBox(
		label,
		widget.NewButton("Telepítés", func() {
			installingInProgress(&w)
		}),
		widget.NewButton("Mégsem", func() {
			a.Quit()
		})))

	w.Resize(fyne.NewSize(300, 200))

	w.ShowAndRun()
}

func installingInProgress(window *fyne.Window) {

	(*window).Resize(fyne.NewSize(300, 190))

	label1 := widget.NewLabelWithStyle(" \n \nTelepítés!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label2 := widget.NewLabelWithStyle("Felkészülés...\n \n ", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	progress := widget.NewProgressBar()

	(*window).SetContent(widget.NewVBox(
		label1,
		label2,
		progress,
	))

	go func() {
		var fileInfo, err = os.Stat(*data + "\\.skyvillage")
		if err != nil || !fileInfo.IsDir() {
			error := os.Mkdir(*data+"\\.skyvillage", os.ModeDir)
			if error != nil {
				log.Fatalln("Can not create directory .skyvillage")
				showErrorScreen(window)
			}
		}
		pBar = progress
		progress.SetValue(0)
		label2.SetText("Java telepítése...\nEz eltarthat pár percig. \n ")
		downloadJava()
		label2.SetText("Launcher telepítése...\n \n ")
		downloadLauncher()
		label2.SetText("Ikon létrehozása...\n \n ")
		downloadFile(*data+"\\.skyvillage\\icon.ico", "https://bendimester23.tk/assets/icon.ico", 0, 0.9)
		downloadFile(*data+"\\.skyvillage\\shortcut.js", "https://gist.githubusercontent.com/Bendimester23/1127a4105d50c68be610d548bef2bec3/raw/4c5e3edef1595e21f7e29ddd47a9e67251b6d14a/makeshortcut.js", 1, 0.9)
		cmd := exec.Command("wscript", *data+"\\.skyvillage\\shortcut.js")
		erro := cmd.Run()
		if erro != nil {
			log.Println(err)
		}
		rErro := os.Remove(*data + "\\.skyvillage\\shortcut.js")
		if rErro != nil {
			log.Println("Error deleting file")
		}
		label2.SetText("Kész!\n \n ")
		showDoneScreen(window)
	}()
}

func showDoneScreen(window *fyne.Window) {
	label1 := widget.NewLabelWithStyle(" \n \nA telepítés sikeres!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label2 := widget.NewLabelWithStyle("Most már bezárhetod ezt az ablakot...\n \n ", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	closeBtn := widget.NewButton("Bezárás", func() {
		a.Quit()
	})
	(*window).SetContent(widget.NewVBox(
		label1,
		label2,
		closeBtn,
	))
	(*window).Resize(fyne.NewSize(300, 185))
}

func showErrorScreen(window *fyne.Window) {
	label1 := widget.NewLabelWithStyle(" \n \nA telepítés sikertelen!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label2 := widget.NewLabelWithStyle("Vedd fel velünk a kapcsolatot...\n \n ", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	closeBtn := widget.NewButton("Bezárás", func() {
		a.Quit()
	})
	(*window).SetContent(widget.NewVBox(
		label1,
		label2,
		closeBtn,
	))
	(*window).Resize(fyne.NewSize(300, 185))
}

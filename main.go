package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
)

var (
	data *string
	a    fyne.App
)

func main() {
	log.Println("Launching SkyLauncher")
	appdata := os.Getenv("APPDATA")

	data = &appdata

	if exists(appdata + "\\.skyvillage\\launcher.jar") {
		launch()
	} else {
		launchInstaller()
	}
	//go startInstaller()
	//startSplash()
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
	checkLauncher()
}

func launch() {
	proc := exec.Command(*data+"\\.skyvillage\\java\\jre\\bin\\java.exe", "-cp", *data+"\\.skyvillage\\launcher.jar", "hu.bendi.skylauncher.Launcher")
	env := proc.Env
	proc.Dir = *data + "\\.skyvillage"
	env = append(proc.Env, "JAVA_HOME="+*data+"\\.skyvillage\\java")
	env = append(env, "APPDATA="+*data)
	proc.Env = env

	proc.Stdout = os.Stdout

	err := proc.Start()

	if err != nil {
		log.Fatalf("Error: %e", err.Error())
	}
	os.Exit(0)
}

func checkLauncher() {
	data := os.Getenv("APPDATA")
	if !exists(data + "\\.skyvillage\\launcher.jar") {
		downloadLauncher()
	}
	launch()
}

func startSplash() {
	a = app.NewWithID("hu.bendi.skylauncher")

	log.Println("Starting splash screen...")
	drv := a.Driver().(desktop.Driver)
	w := drv.CreateSplashWindow()

	w.SetFixedSize(true)
	w.CenterOnScreen()

	w.SetContent(widget.NewLabel("SkyLauncher\n   Loading..."))
	w.ShowAndRun()
	defer w.Hide()
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
				log.Fatal(err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func downloadLauncher() {
	downloadFile(*data+"\\.skyvillage\\launcher.jar", "https://www.dropbox.com/s/jki9u6ecbxzu6er/skylauncher-0.1-SNAPSHOT.jar?dl=1")
}

func downloadJava() {
	log.Println("Downloading java...")
	url := "https://www.dropbox.com/s/j9wniroiggs5vkn/java.zip?dl=1"
	erro := mdIfNotPresent(*data + "\\.skyvillage\\tmp")
	if erro != nil {
		log.Fatalln("Error creating directory")
		log.Fatalf("Error: %s", erro)
		os.Exit(500)
	}
	err := downloadFile(*data+"\\.skyvillage\\tmp\\java.zip", url)
	if err != nil {
		log.Fatalf("Error: %s", err)
		log.Fatalln("Error while downloading java. Please contact support.")
	}
	log.Println("Installing java...")
	unzip(*data+"\\.skyvillage\\tmp\\java.zip", *data+"\\.skyvillage\\")
	rErro := os.Remove(*data + "\\.skyvillage\\tmp\\java.zip")
	if rErro != nil {
		log.Fatalln("Error deleting file")
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

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
func checkJava() {
	data := os.Getenv("APPDATA")
	_, err := os.Stat(data + "\\.skyvillage\\java\\bin\\java.exe")
	if err != nil {
		log.Println("Java not found, downloading...")
		downloadJava()
	} else {
		proc := exec.Command(data+"\\.skyvillage\\java\\bin\\java.exe", "--version")
		erro := proc.Start()
		if erro != nil {
			log.Println("Java corrupted, downloading...")
			downloadJava()
		}
	}
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

	label := widget.NewLabelWithStyle(" \n \n \nTelepítés!\n \n \n ", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	progress := widget.NewProgressBar()

	(*window).SetContent(widget.NewVBox(
		label,
		progress))
}

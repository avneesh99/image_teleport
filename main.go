package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ImageManifest struct {
	Layers []string `json:"Layers"`
}

const reconstructorScript = `#!/bin/bash
# Create the new image tar
tar -cf reconstructed_image.tar image_layers

# Load the new image
docker load -i ../reconstructed_image.tar

# List the images
docker images

echo "Docker image successfully reconstructed and loaded!"
`

func main() {
	imageName := flag.String("image", "", "Docker image name and tag (e.g., your_image_name:tag)")
	remoteHost := flag.String("remote", "", "Remote host (e.g., ec2-user@10.0.140.24)")
	bastionHost := flag.String("bastion", "", "Bastion host (optional, e.g., ec2-user@ec2-13-127-245-236.ap-south-1.compute.amazonaws.com)")
	identityFile := flag.String("identity", "", "Path to SSH identity file")
	destinationPath := flag.String("dest", "/tmp/docker_layers", "Destination path on remote host")

	flag.Parse()

	if *imageName == "" || *remoteHost == "" || *identityFile == "" {
		fmt.Println("Error: image, remote, and identity flags are required")
		flag.Usage()
		os.Exit(1)
	}

	println("Creating docker_layers directory")
	tempDir, err := os.MkdirTemp("", "docker_layers")
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		os.Exit(1)
	}
	println("Created docker_layers directory @ ", tempDir)

	defer os.RemoveAll(tempDir)

	tarFile := filepath.Join(tempDir, "image.tar")
	println("Running ", "docker", "save", "-o", tarFile, *imageName)
	err = runCommand("docker", "save", "-o", tarFile, *imageName)
	if err != nil {
		fmt.Printf("Error saving Docker image: %v\n", err)
		os.Exit(1)
	}
	println("Done: ", "docker", "save", "-o", tarFile, *imageName)

	layersDir := filepath.Join(tempDir, "image_layers")
	err = os.Mkdir(layersDir, 0755)
	if err != nil {
		fmt.Printf("Error creating layers directory: %v\n", err)
		os.Exit(1)
	}

	println("Running ", "tar", "-xf", tarFile, "-C", layersDir)
	err = runCommand("tar", "-xf", tarFile, "-C", layersDir)
	if err != nil {
		fmt.Printf("Error extracting layers: %v\n", err)
		os.Exit(1)
	}
	println("Done: ", "tar", "-xf", tarFile, "-C", layersDir)

	layers, err := getImageLayers(layersDir)
	if err != nil {
		fmt.Printf("Error getting image layers: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Layers: ", layers)

	reconstructorPath := filepath.Join(tempDir, "reconstruct.sh")
	err = os.WriteFile(reconstructorPath, []byte(reconstructorScript), 0755)
	if err != nil {
		fmt.Printf("Error creating reconstructor script: %v\n", err)
		os.Exit(1)
	}

	rsyncArgs := []string{"-avzc", "--delete"}
	sshCommand := fmt.Sprintf("ssh -i %s", *identityFile)
	if *bastionHost != "" {
		sshCommand += fmt.Sprintf(" -o ProxyJump=%s", *bastionHost)
	}
	rsyncArgs = append(rsyncArgs, "-e", sshCommand)
	rsyncArgs = append(rsyncArgs, layersDir, reconstructorPath)
	rsyncArgs = append(rsyncArgs, fmt.Sprintf("%s:%s", *remoteHost, *destinationPath))

	fmt.Println("Running command: ", "rsync", rsyncArgs)
	err = runCommand("rsync", rsyncArgs...)
	if err != nil {
		fmt.Printf("Error pushing layers and reconstructor to remote host: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Changed Docker image layers, reconstructor script, and layers JSON pushed successfully!")
	fmt.Println("\nTo reconstruct the image on the remote host, run:")
	fmt.Printf("sudo ./reconstruct.sh \n")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getImageLayers(layersDir string) ([]string, error) {
	manifestPath := filepath.Join(layersDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest: %v", err)
	}

	var manifests []ImageManifest
	err = json.Unmarshal(manifestData, &manifests)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling manifest: %v", err)
	}

	if len(manifests) == 0 {
		return nil, fmt.Errorf("no manifests found")
	}

	return manifests[0].Layers, nil
}

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts         []string `yaml:"hosts"`
	Image         string   `yaml:"image"`
	ContainerPort int      `yaml:"containerPort"`
}

var (
	environment      *string
	sshUsername      *string
	registryName     *string
	registryUsername *string
	sshPort          *int
	path             *string
	verbose          *bool
)

func init() {
	environment = flag.String("environment", "", "The environment to deploy")
	sshUsername = flag.String("sshUsername", "", "The SSH username to use")
	sshPort = flag.Int("sshPort", 22, "The SSH port to use") //nolint:gomnd
	registryName = flag.String("registryName", "ghcr.io", "The name of the registry to use")
	registryUsername = flag.String("registryUsername", "", "The username of the registry to use")
	path = flag.String("path", "deployments", "The path to the deployment files")
	verbose = flag.Bool("verbose", false, "Enable verbose logging")
}

func main() { //nolint:cyclop // TODO(joelrose): refactor
	flag.Parse()

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if *verbose {
		log = log.Level(zerolog.DebugLevel)
	}

	if *environment == "" || *registryUsername == "" || *sshUsername == "" {
		log.Info().Msg("Missing required arguments: -environment, -registryUsername, -sshUsername")
		flag.PrintDefaults()
		os.Exit(1)
	}

	privateKey := os.Getenv("SSH_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal().Msg("SSH_PRIVATE_KEY not set")
	}

	registryPassword := os.Getenv("REGISTRY_PASSWORD")

	ssh, err := NewSSH(*sshUsername, privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create ssh client")
	}

	path := fmt.Sprintf(*path+"/%s.*.yml", *environment)
	files, err := filepath.Glob(path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find deployment files")
	}

	for _, file := range files {
		containerName := strings.Split(filepath.Base(file), ".")[1]

		log = log.With().Str("file", file).Str("container", containerName).Logger()

		f, err := os.ReadFile(file)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read config")
		}

		var config Config
		err = yaml.Unmarshal(f, &config)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to unmarshal config file")
		}

		// TODO(joelrose): this could panic
		desiredImageHash := strings.Split(config.Image, ":")[1]

		log.Debug().Str("desiredImageHash", desiredImageHash).Msg("deploying...")

		for _, host := range config.Hosts {
			addr := host + ":" + strconv.Itoa(*sshPort)

			log.Debug().Str("addr", addr).Msg("connecting...")

			tmpl, err := RenderTemplate(TemplateData{
				RegistryName:     *registryName,
				RegistryUsername: *registryUsername,
				RegistryPassword: registryPassword,
				DesiredImageHash: desiredImageHash,
				ContainerName:    containerName,
				Image:            config.Image,
				Host:             host,
				Environment:      *environment,
				ContainerPort:    config.ContainerPort,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to render template")
			}

			_, err = ssh.RunCommand(addr, tmpl)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to run command")
			}
		}
	}
}

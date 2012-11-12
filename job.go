package main

import (
	"bytes"
	"fmt"
	"github.com/rdwilliamson/aws/glacier"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func job(args []string) {
	if len(args) < 1 {
		fmt.Println("no job command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "inventory":
		args = getConnection(args)

		if len(args) < 1 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]
		args = args[1:]

		var description, topic string
		if len(args) > 0 {
			topic = args[0]
		}
		if len(args) > 1 {
			description = args[1]
		}

		jobId, err := connection.InitiateInventoryJob(vault, topic, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(jobId)

	case "archive":
		args = getConnection(args)

		if len(args) < 2 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]
		archive := args[1]
		args = args[2:]

		var description, topic string
		if len(args) > 0 {
			topic = args[0]
		}
		if len(args) > 1 {
			description = args[1]
		}

		jobId, err := connection.InitiateRetrievalJob(vault, archive, topic, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(jobId)

	case "list":
		args = getConnection(args)

		if len(args) < 1 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]

		jobs, _, err := connection.ListJobs(vault, "", "", "", 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, v := range jobs {
			fmt.Println("Action:", v.Action)
			if v.Action == "ArchiveRetrieval" {
				fmt.Println("Archive ID:", v.ArchiveId)
				fmt.Println("Archive Size:", v.ArchiveSizeInBytes, prettySize(v.ArchiveSizeInBytes))
			}
			fmt.Println("Completed:", v.Completed)
			if v.Completed {
				fmt.Println("Completion Date:", v.CompletionDate)
			}
			fmt.Println("Creation Date:", v.CreationDate)
			if v.Completed && v.Action == "InventoryRetrieval" {
				fmt.Println("Invenotry Size:", v.InventorySizeInBytes, prettySize(uint64(v.InventorySizeInBytes)))
			}
			fmt.Println("Job Description:", v.JobDescription)
			fmt.Println("Job ID:", v.JobId)
			if v.Action == "ArchiveRetrieval" {
				fmt.Println("SHA256 Tree Hash:", v.SHA256TreeHash)
			}
			fmt.Println("SNS Topic:", v.SNSTopic)
			fmt.Println("Status Code:", v.StatusCode)
			fmt.Println("Status Message:", v.StatusMessage)
			fmt.Println("Vault ARN:", v.VaultARN)
			fmt.Println()
		}

	case "describe":
		args = getConnection(args)

		if len(args) < 2 {
			fmt.Println("no vault and/or job id")
			os.Exit(1)
		}
		vault := args[0]
		jobId := args[1]

		job, err := connection.DescribeJob(vault, jobId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Action:", job.Action)
		if job.Action == "ArchiveRetrieval" {
			fmt.Println("Archive ID:", job.ArchiveId)
			fmt.Println("Archive Size:", job.ArchiveSizeInBytes, prettySize(job.ArchiveSizeInBytes))
		}
		fmt.Println("Completed:", job.Completed)
		if job.Completed {
			fmt.Println("Completion Date:", job.CompletionDate)
		}
		fmt.Println("Creation Date:", job.CreationDate)
		if job.Completed && job.Action == "InventoryRetrieval" {
			fmt.Println("Invenotry Size:", job.InventorySizeInBytes, prettySize(uint64(job.InventorySizeInBytes)))
		}
		fmt.Println("Job Description:", job.JobDescription)
		fmt.Println("Job ID:", job.JobId)
		if job.Action == "ArchiveRetrieval" {
			fmt.Println("SHA256 Tree Hash:", job.SHA256TreeHash)
		}
		fmt.Println("SNS Topic:", job.SNSTopic)
		fmt.Println("Status Code:", job.StatusCode)
		fmt.Println("Status Message:", job.StatusMessage)
		fmt.Println("Vault ARN:", job.VaultARN)

	case "get":
		if len(args) < 1 {
			fmt.Println("no job sub command")
			os.Exit(1)
		}
		subCommand := args[0]
		args = args[1:]

		switch subCommand {
		case "inventory":
			args = getConnection(args)

			if len(args) < 2 {
				fmt.Println("no vault and/or job id")
				os.Exit(1)
			}
			vault := args[0]
			job := args[1]

			inventory, err := connection.GetInventoryJob(vault, job)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println("Vault ARN:", inventory.VaultARN)
			fmt.Println("Inventory Date:", inventory.InventoryDate)

			for _, v := range inventory.ArchiveList {
				fmt.Println()
				fmt.Println("Archive ID:", v.ArchiveId)
				fmt.Println("Archive Description:", v.ArchiveDescription)
				fmt.Println("Creation Date:", v.CreationDate)
				fmt.Println("Size:", v.Size, prettySize(v.Size))
				fmt.Println("SHA256 Tree Hash:", v.SHA256TreeHash)
			}

		case "archive":
			args = getConnection(args)

			if len(args) < 3 {
				fmt.Println("no vault, job id, and/or output file")
				os.Exit(1)
			}
			vault := args[0]
			job := args[1]
			fileName := args[2]

			file, err := os.Create(fileName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer file.Close()

			archive, _, err := connection.GetRetrievalJob(vault, job, 0, 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer archive.Close()

			_, err = io.Copy(file, archive)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown job sub command:", subCommand)
			os.Exit(1)
		}

	case "run":
		args = getConnection(args)
		if len(args) < 4 {
			fmt.Println("no vault, archive, download size and/or output file")
			os.Exit(1)
		}
		vault := args[0]
		archive := args[1]
		partSize, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		partSize *= 1024 * 1024
		output := args[3]
		args = args[3:]

		var topic string
		if len(args) > 0 {
			topic = args[0]
			args = args[1:]
		}

		var description string
		if len(args) > 0 {
			description = args[0]
			args = args[1:]
		}

		// initiate retrieval job
		job, err := connection.InitiateRetrievalJob(vault, archive, topic, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Println("initiated retrieval job:", job)

		// wait for job to complete, using polling
		time.Sleep(3 * time.Hour)

		// check status sleeping 15m?
		try := 0
		var size uint64
		var fullTreeHash string
		for {
			job, err := connection.DescribeJob(vault, job)
			if err != nil {
				log.Println(err)
				try++
				if try > retries {
					fmt.Println("too many retries")
					os.Exit(1)
				}
			} else {
				try = 0
				if job.Completed {
					size = uint64(job.ArchiveSizeInBytes)
					fullTreeHash = job.SHA256TreeHash
					log.Println("job completed")
					log.Println("size:", prettySize(size))
					log.Println("tree hash:", fullTreeHash)
					break
				}
				log.Println("retrieval job not yet completed")
				time.Sleep(15 * time.Minute)
			}
		}

		log.Println("creating file", output)

		file, err := os.OpenFile(output, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		// loop getting parts, checking tree hash of each
		bufferData := make([]byte, 0, partSize)
		buffer := bytes.NewBuffer(bufferData)
		log.Println("downloading in", prettySize(partSize), "chunks,", size/partSize+1, "parts", partSize)
		n := uint64(0)
		hasher := glacier.NewTreeHash()

		for n < size {
			log.Println("downloading", n, "to", n+partSize-1)
			part, treeHash, err := connection.GetRetrievalJob(vault, job, n, n+partSize-1)
			if err != nil {
				log.Println("GetRetrievalJob:", err)
				try++
				if try > retries {
					fmt.Println("too many retries")
					os.Exit(1)
				}
				continue
			}

			log.Println("starting download of part")
			// copy to temporary buffer
			nn, err := io.Copy(buffer, part)
			if err != nil {
				log.Println("io.Copy:", err)
				try++
				if try > retries {
					fmt.Println("too many retries")
					os.Exit(1)
				}
				continue
			}
			log.Println("copied", nn, "bytes to buffer")

			// check tree hash
			hasher.Write(buffer.Bytes())
			hasher.Close()
			if treeHash != hasher.TreeHash() {
				log.Println("tree hash mismatch")
				log.Println("wanted", treeHash)
				log.Println("got", hasher.TreeHash())
				try++
				if try > retries {
					fmt.Println("too many retries")
					os.Exit(1)
				}
				continue
			}
			log.Println("checked tree hash")

			// copy to file
			_, err = file.Write(buffer.Bytes())
			if err != nil {
				log.Println("copying buffer to file:", err)
				try++
				if try > retries {
					fmt.Println("too many retries")
					os.Exit(1)
				}
			}
			log.Println("copied to file")

			n += uint64(buffer.Len())
			try = 0
			buffer.Reset()
			hasher.Reset()
		}

		log.Println("download complete, checking ")
		// check tree hash of entire archive
		log.Println(fullTreeHash)
		_, err = file.Seek(0, 0)
		if err != nil {
			log.Println("seek:", err)
			os.Exit(1)
		}

		_, err = io.Copy(hasher, file)
		if err != nil {
			log.Println("hashing whole file:", err)
			os.Exit(1)
		}
		hasher.Close()

		if hasher.TreeHash() != fullTreeHash {
			log.Println("entire file tree hash mismatch")
			os.Exit(1)
		}

	default:
		fmt.Println("unknown job command:", command)
		os.Exit(1)
	}
}

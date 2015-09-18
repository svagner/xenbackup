package main

import (
	"flag"
	"log"

	"github.com/svagner/xenbackup/xen"
)

var host = flag.String("host", "", "Hypervisor's name")
var username = flag.String("user", "", "Hypervisor's username")
var password = flag.String("pass", "", "Hypervisor's password")
var vmname = flag.String("vmname", "", "Virtual machine's name")
var vmuuid = flag.String("vmuuid", "", "Virtual machine's uuid")
var path = flag.String("path", "", "Path for save backups")

func main() {
	flag.Parse()
	var backupXen xen.XenBackup
	err := backupXen.Init(*host, *username, *password)
	if err != nil {
		log.Fatalln(err)
	}
	backupXen.VMs = append(backupXen.VMs, xen.VM{Name: *vmname, Uuid: *vmuuid})
	err = backupXen.Backup(*path)
	if err != nil {
		log.Fatalln("Error", err)
	}
	log.Println("Finished")
}

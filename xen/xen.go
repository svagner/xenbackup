package xen

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"time"

	xapi "github.com/svagner/go-xenserver-client"
)

const (
	EXPORT_SCHEMA    = "http://"
	EXPORT_METHOD    = "GET"
	EXPORT_EXTENTION = ".xva"
	SNAPSHOT_PREFIX  = "backup_"
)

type VM struct {
	Name string
	Uuid string
}

type authData struct {
	host   string
	user   string
	passwd string
}

type XenBackup struct {
	client xapi.XenAPIClient
	auth   authData
	VMs    []VM
}

func (xb *XenBackup) Init(host, user, pass string) (err error) {
	xb.client = xapi.NewXenAPIClient(host, user, pass)
	err = xb.client.Login()
	if err != nil {
		return
	}
	xb.auth.host = host
	xb.auth.passwd = pass
	xb.auth.user = user
	return
}

func (xb *XenBackup) generateBackupName(vmname string) string {
	now := time.Now()
	return SNAPSHOT_PREFIX + vmname + "_" + fmt.Sprintf("%d%d%d%d%d", now.Year(),
		now.Month(), now.Day(), now.Hour(), now.Minute())

}

func (xb *XenBackup) Backup(path string) error {
	vms := make([]*xapi.VM, 0)
	for _, vm := range xb.VMs {
		if vm.Uuid != "" {
			node, err := xb.client.GetVMByUuid(vm.Uuid)
			if err != nil {
				return err
			}
			vms = append(vms, node)
			continue
		}
		if vm.Name != "" {
			res, err := xb.client.GetVMByNameLabel(vm.Name)
			if err != nil {
				return err
			}
			if len(res) == 0 {
				return errors.New("VM not found by name " + vm.Name)
			}
			for _, node := range res {
				vms = append(vms, node)
			}
		}
	}
	// Start doing snapshots
	for _, vm := range vms {
		name, err := vm.GetName()
		if err != nil {
			return err
		}

		backupName := xb.generateBackupName(name)
		log.Println("Start to create snapshot", backupName)
		snapshot, err := vm.Snapshot(backupName)
		if err != nil {
			return err
		}
		err = snapshot.SetTemplateAsVM(false)
		if err != nil {
			return err
		}
		uuid, err := snapshot.GetUuid()
		if err != nil {
			return err
		}
		log.Println("export snapshot", backupName, "to file", path+"/"+backupName+EXPORT_EXTENTION)
		err = xb.exportBackup(uuid, path+"/"+backupName+EXPORT_EXTENTION)
		if err != nil {
			return err
		}
		//snapshotName, err := snapshot.GetName()
		snapshotName, err := snapshot.GetName()
		if err != nil {
			return err
		}
		if strings.HasPrefix(snapshotName, SNAPSHOT_PREFIX) {
			log.Println("Trying to delete snapshot ", snapshotName)
			log.Println("Get snapshot's list:")
			disks, err := snapshot.GetDisks()
			if err != nil {
				return errors.New("Get disk for snapshot error:" + err.Error())
			}
			log.Println("Snapshot", snapshotName, "disks:")
			for idx, disk := range disks {
				u, e := disk.GetUuid()
				if e != nil {
					return errors.New("Get disk uuid for snapshot error:" + err.Error())
				}
				log.Println("UUID", idx, u)
			}
			// first step - delete snapshot
			err = snapshot.Destroy()
			if err != nil {
				return errors.New("Error while destroy snapshot " + backupName + ": " + err.Error())
			}
			// as second step - we have to delete vdi drive os snapshot
			log.Println("Delete vdi's for snapshot " + backupName)
			for idx, disk := range disks {
				u, e := disk.GetUuid()
				if e != nil {
					return errors.New("Get disk uuid for delete snapshot error:" + err.Error())
				}
				log.Println("Deleting vdi ", idx, u)
				e = disk.Destroy()
				if e != nil {
					return errors.New("Delete vdi" + u + "failed:" + e.Error())
				}
			}

		} else {
			return errors.New("VM image wasn't delete: name '" + name + "' hasn't prefix '" + SNAPSHOT_PREFIX + "'")
		}
	}
	return nil
}

func (xb *XenBackup) exportBackup(uuid, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()
	client := &http.Client{}
	req, err := http.NewRequest(EXPORT_METHOD, EXPORT_SCHEMA+xb.auth.host+"/export?uuid="+uuid+"&format=raw", nil)
	req.SetBasicAuth(xb.auth.user, xb.auth.passwd)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

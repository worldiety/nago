---
title: "Backup Management"
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/backup_management/galleries/overview/overview.png"
  - src: "/images/systems/backup_management/galleries/overview/backup.png"
  - src: "/images/systems/backup_management/galleries/overview/recovery.png"
  - src: "/images/systems/backup_management/galleries/overview/export_master_key.png"
  - src: "/images/systems/backup_management/galleries/overview/new_master_key.png"
---

The Backup Management system provides full backup and restore capabilities for the application. It also allows management of the Nago master key, which is used for encrypting sensitive stores (e.g., sessions, secrets).

## Functional areas
Backup Management provides the following key functions:

### Backup
Creates a complete backup of the application data. **Encrypted stores** remain encrypted and cannot be restored without the **master key**. The backup file is downloaded as a ZIP.

{{< callout type="warning" >}}
Ensure that no operations are performed in parallel to maintain a consistent backup.
{{< /callout >}}

### Restore
Restores the application state from a backup file. All existing data in the restored stores will be overwritten. Only backup files from **trusted sources** should be used.

{{< callout type="warning" >}}
Encrypted stores require the master key to be restored correctly.
{{< /callout >}}

### Export Master Key
Allows the export of the **Nago master key**. This key is required to decrypt encrypted stores in backups (e.g., sessions or secrets).

{{< callout type="warning" >}}
Handle the key securely. If exposed, all encrypted data is considered compromised.
{{< /callout >}}

### Replace Master Key
Allows replacing the current **Nago master key** with a new one. All encrypted stores will then require the new key for decryption.

{{< callout type="info" >}}
The service must be restarted for the new key to take effect.
{{< /callout >}}

{{< swiper name="galleryOverview" loop="false" >}}

## Dependencies
**Requires:**
- No other systems

**Is required by:**
- none

## Activation
This system is activated via:
```go
std.Must(cfg.BackupManagement())
```

```go
backupManagement := std.Must(cfg.BackupManagement())
```
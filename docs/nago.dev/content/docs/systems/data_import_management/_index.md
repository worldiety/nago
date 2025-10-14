---
title: Data Import Management
galleryOverview:
- src: "/images/systems/shared/admin_center.png"
- src: "/images/systems/data_import_management/galleries/overview/admin_center.png"
- src: "/images/systems/data_import_management/galleries/overview/user_importer.png"
galleryWorkflow:
- src: "/images/systems/data_import_management/galleries/workflow/new_user_import.png"
- src: "/images/systems/data_import_management/galleries/workflow/data_uploaded.png"
- src: "/images/systems/data_import_management/galleries/workflow/mapping_schema_1.png"
- src: "/images/systems/data_import_management/galleries/workflow/mapping_schema_2.png"
- src: "/images/systems/data_import_management/galleries/workflow/mapped_data.png"
- src: "/images/systems/data_import_management/galleries/workflow/validation_and_transformation.png"
- src: "/images/systems/data_import_management/galleries/workflow/import_options.png"
- src: "/images/systems/data_import_management/galleries/workflow/import_status.png"
---


**Data Import Management** provides tools to import structured and semi-structured data into the Nago ecosystem.  
It allows mapping uploaded data to existing internal entities, such as users, enabling data reuse and synchronization across systems.  
Therefore, it supports multiple data formats such as **CSV**, **JSON** and **PDF AcroForms**, and provides interactive tools for reviewing, transforming and importing data. 

Data Import Management is designed for administrators and power users who need to import, validate, and align data with the internal structures of the Nago application.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Data Import Management provides the following key functions:

### Data staging and review
- Upload and stage data from supported formats
- Preview parsed records before importing
- Identify and resolve potential validation issues early
- Available data formats depend on which **parsers** are activated in the configuration

{{< callout type="info" >}}
Parsers are modular components that must be explicitly enabled by the developer.  
This allows the application to control which input formats (e.g., CSV, JSON, PDF AcroForms) are supported.
{{< /callout >}}

Parsers can be activated via:
```go
option.MustZero(imports.UseCases.RegisterParser(user.SU(), csv.NewParser()))
```

### Field mapping
- Define a dedicated import schema for each uploaded file
- Map imported data fields to existing entity attributes
- Automatically detect matching fields based on header names or structure

### Validation and transformation
- Perform manual validation of imported data
- Apply custom transformation logic to adapt input before importing
- Detect conflicts (e.g., duplicates or missing required fields)
- Track import progress

### Import execution
- Execute imports into existing repositories
- Review imported entities or error logs directly within the Admin Center

## Functional Flow
1. **Select Importer and Format**  
   Users choose the import type (e.g., user import) and file format (CSV, JSON, PDF).

2. **Upload File**  
   Depending on the parser configuration, supported file types can be uploaded directly.

3. **Field Mapping**  
   The uploaded data is automatically mapped to existing entity structures.
   Mappings can be reviewed, adjusted, or manually filled in for missing fields.

4. **Review Entries**  
   Each imported entry can be viewed side by side:
   - Raw input data
   - Transformed application entity  
     Users can confirm or reject entries, navigate through records, and monitor progress.

5. **Import Execution**  
   Confirmed entries are imported into the system.  
   Options include continuing on errors and merging duplicates.  
   Merge behavior can be customized (e.g., whether new values override existing ones).

{{< swiper name="galleryWorkflow" loop="false" >}}

## Extensibility
{{< callout type="info" >}}
The system is **extensible** — developers can register custom parsers and importers  
by implementing the `parser.Parser` and `importer.Importer` interfaces.
{{< /callout >}}

- **Importer Interface** — Defines how parsed data is imported into specific domain entities.
- **Parser Interface** — Defines how raw file content (CSV, JSON, etc.) is parsed into structured objects.



## Dependencies
Data Import Management operates independently and does not depend on other systems.

## Activation
This system is activated via:
```go
importManagement := std.Must(cfgdataimport.Enable(cfg))
```

The importer for user entities is activated via:
```go
userManagementUCs := std.Must(cfg.UserManagement()).UseCases

option.MustZero(imports.UseCases.RegisterImporter(user.SU(), userimporter.NewImporter(userManagementUCs)))
option.MustZero(imports.UseCases.RegisterParser(user.SU(), csv.NewParser()))
option.MustZero(imports.UseCases.RegisterParser(user.SU(), pdf.NewParser()))
option.MustZero(imports.UseCases.RegisterParser(user.SU(), json.NewParser()))
```
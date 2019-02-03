# petscii

Generates the PETSCII tables using the Unicode to PETSCII mapping table compiled by Rebecca Turner. The source used to generate this table are not found in this repository. Download and place in the following location, relative to the repository root:

- ext/petscii/table.txt

Document downloaded from here:

https://github.com/9999years/Unicode-PETSCII/blob/master/table.txt

Generate `petscii.go` with:

```bash
go generate
```


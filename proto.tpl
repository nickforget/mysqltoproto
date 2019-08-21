{{range $key, $value := . }}
message {{$key}}{ {{range $index, $data := $value }}
    optional {{$data.DataType}} {{$data.ColumnName}} = {{add $index}}; // {{$data.ColumnComment}}{{end}}
}
{{end}}
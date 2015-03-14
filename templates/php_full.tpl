<?php{{ .AdditionalDeclaration }}{{ .PrepareBody }}{{ .PrepareHeader }}
$ctx = stream_context_create([
  "http" => [
    "method" => "{{ .Method }}"{{ .Header }}{{ .Content }}
  ]
]);
$fp = fopen({{ .Url }}, "r", false, $ctx);
if ($fp === false)
  exit();
var_dump(stream_get_meta_data($fp));
var_dump(stream_get_contents($fp));

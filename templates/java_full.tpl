{{ range $key, $_ := .Modules }}import {{ $key }};
{{end}}

public class Main { {{ .AdditionalDeclaration }}
    public static void main(String[] args) {
        try {
            {{ .CommonInitialize }}{{ .PrepareBody }}URL url = new URL({{ .Url }});

            {{ .ConnectionClass }} conn = ({{ .ConnectionClass }})url.openConnection({{ .Proxy }});
{{ .PrepareConnection }}
            System.out.printf("Response: %d %s\n", conn.getResponseCode(), conn.getResponseMessage());
            BufferedReader br = new BufferedReader(new InputStreamReader(conn.getInputStream()));
            String input;

            while ((input = br.readLine()) != null) {
                System.out.println(input);
            }
            br.close();
        } catch (MalformedURLException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}

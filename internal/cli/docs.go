package cli

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Documentation and examples for Maniplacer",
	Long:  ``,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetString("port")
		if err != nil {
			fmt.Printf("Could not parse port due to %s, using default...\n", err)
			port = "8080"
		}

		fmt.Printf("Documentation server available in: http://localhost:%s/docs\n", port)

		mux := http.NewServeMux()

		mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(page))
		})

		http.ListenAndServe(fmt.Sprintf(":%s", port), mux)

	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringP("port", "p", "8000", "Port for serving docs page")
}

var page = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Maniplacer Docs</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      max-width: 900px;
      margin: 2rem auto;
      line-height: 1.6;
    }
    h1, h2 {
      color: #2c3e50;
    }
    pre {
      background: #f4f4f4;
      padding: 12px;
      border-radius: 8px;
      overflow-x: auto;
    }
    code {
      font-family: monospace;
    }
    .explanation {
      background: #eef6ff;
      padding: 10px;
      border-left: 4px solid #3498db;
      margin-bottom: 1rem;
    }
    ul {
      background: #fafafa;
      padding: 10px 20px;
      border-left: 4px solid #27ae60;
      border-radius: 6px;
    }
    li {
      margin-bottom: 8px;
    }
  </style>
</head>
<body>
  <h1>📘 Maniplacer Examples</h1>

  <hr>

  <h2>Template example</h2>
  <pre><code>apiVersion: v1
kind: Secret
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
type: Opaque
data:
  {{- range $key, $value := .secrets }}
  {{ $key }} : {{ $value | Base64 | Quote }}
  {{- end }}</code></pre>

  <div class="explanation">
    <p><b>Why <code>range</code>?</b></p>
    <p>
      The <code>range</code> command loops over a collection.  
      In this case, <code>.secrets</code> is a map of key–value pairs.  
      Each iteration assigns the map key to <code>$key</code> and the value to <code>$value</code>,  
      allowing us to output each secret entry as YAML.
    </p>

    <p><b>What does the <code>|</code> operator do?</b></p>
    <p>
      The pipe operator <code>|</code> passes the output of one function into another, like in Unix shells.  
      For example:  
      <code>{{ $value | Base64 | Quote }}</code>  
      means:  
      1. Take <code>$value</code>  
      2. Encode it with <code>Base64</code>  
      3. Then pass the result into <code>Quote</code> to wrap it in quotes.  
    </p>
  </div>

  <hr>

  <h2>⚙️ Built-in Functions</h2>
  <ul>
    <li><b><code>Base64</code></b> – Encodes a string into Base64.  
      <br><i>Example:</i> <code>{{ "hello" | Base64 }}</code> → <code>aGVsbG8=</code></li>

    <li><b><code>ToUpper</code></b> – Converts text to uppercase.  
      <br><i>Example:</i> <code>{{ "maniplacer" | ToUpper }}</code> → <code>MANIPLACER</code></li>

    <li><b><code>ToLower</code></b> – Converts text to lowercase.  
      <br><i>Example:</i> <code>{{ "KUBERNETES" | ToLower }}</code> → <code>kubernetes</code></li>

    <li><b><code>Quote</code></b> – Wraps text in quotes.  
      <br><i>Example:</i> <code>{{ "world" | Quote }}</code> → <code>"world"</code></li>
  </ul>

</body>
</html>
`

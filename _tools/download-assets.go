/* Script to download assets.
 * Usage:
 * $ go run ./_tools/download-assets.go */

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	getAsset(
		"https://github.githubassets.com/favicons/favicon.svg",
		"./internal/server/static/favicon.svg",
		// since the URL is unversioned it is expected that this may change in future
		"6a9577cd4f7fa6b75bde1025af85b944e9dd1388373b55ccba6e9f80ac2eae60",
	)
	getAsset(
		"https://raw.githubusercontent.com/sindresorhus/github-markdown-css/refs/tags/v5.8.1/github-markdown-dark.css",
		"./internal/server/static/github-markdown-dark.css",
		"a147b7b29753ef78c807d3b7921de2eb9f9165c59b16db3848236a5599f50f1b",
	)
	getAsset(
		"https://raw.githubusercontent.com/sindresorhus/github-markdown-css/refs/tags/v5.8.1/github-markdown-light.css",
		"./internal/server/static/github-markdown-light.css",
		"a1a198514565120cb1660fcb4583e3eaa00d84294ef8cf989d6c6aa7ffc0e1c0",
	)
	getAsset(
		"https://cdnjs.cloudflare.com/ajax/libs/mermaid/11.3.0/mermaid.min.js",
		"./internal/server/static/mermaid.min.js",
		"0d2b6f2361e7e0ce466a6ed458e03daa5584b42ef6926c3beb62eb64670ca261",
	)
	getAsset(
		"https://cdnjs.cloudflare.com/ajax/libs/mathjax/3.2.2/es5/tex-mml-chtml.min.js",
		"./internal/server/static/tex-mml-chtml.min.js",
		"11d5c53772a9eb9b95a26d6df0cd8bc4b93d6e305dea4407e611f7e2a7c68657",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_AMS-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_AMS-Regular.woff",
		"3de784d07b9fa8f104c10928a878ee879cf3305cae5195cba663c9c2bb0195eb",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Calligraphic-Bold.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Calligraphic-Bold.woff",
		"af04542b29eaac04550a140c5f1760a649783989426f2540855bf4157819367d",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Calligraphic-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Calligraphic-Regular.woff",
		"26683bf201fb258a2237d9754616de9d4ecf4cc1cd39dd1902476df7d75f1d16",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Fraktur-Bold.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Fraktur-Bold.woff",
		"721921bab0d001ebff0206c24cb5de6ca136467bf1843a7aa32030ba061d1e92",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Fraktur-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Fraktur-Regular.woff",
		"870673df72e70f87c91a5a317d558c2c3b54392264ad79bbadc6d424ae8765fe",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Main-Bold.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Main-Bold.woff",
		"88b98cad3688915e50da55553ff6ad185e0dce134b47f176e91b100f8a9b175c",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Main-Italic.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Main-Italic.woff",
		"355254db9ca10a09a3b5f0929d74eb4670f44fbe864c07526a06213e0a0caf6c",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Main-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Main-Regular.woff",
		"1cb1c39ea642f26a4dfed230b4aea1c3c218689421f6e9c0a7c1811693c4fa07",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Math-BoldItalic.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Math-BoldItalic.woff",
		"8ea8dbb1b02e6f730f55b4cb5d413b785b9f5c39807d0a0fa7da206a0824a457",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Math-Italic.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Math-Italic.woff",
		"a009bea404f7a500ded48f8b9ad9cf16e12504b3195dd9e25975289b8256b0f0",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Math-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Math-Regular.woff",
		"c01d3321e89b403c4b811aa153c4e618eda3421f92d8a072a02c8d190782a191",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_SansSerif-Bold.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_SansSerif-Bold.woff",
		"32792104b5ef69eded905b6d1598ed05d8087684d38e7a94d52e3c38ba16f47e",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_SansSerif-Italic.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_SansSerif-Italic.woff",
		"fc6ddf5df402b263cfb158aed8e89972542c34b719cd87b1db30461985f7bd5b",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_SansSerif-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_SansSerif-Regular.woff",
		"b418136e3b384baaadecb70bd3c48a8da9825210130b2897db808c44efc883cb",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Script-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Script-Regular.woff",
		"af96f67d7accf5fd2a4a682d0b9f8b339f8ea6fe34c310c1694c8ba7f6ddc96f",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Size1-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Size1-Regular.woff",
		"c49810b53ecc0d87d8028762c518924197dd9d3f905b08f99ea241301085b9cb",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Size2-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Size2-Regular.woff",
		"30e889b58cbc51adfbb038ab1a96dc4025aa3542a3cff7712fc55ece510675e2",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Size3-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Size3-Regular.woff",
		"5cda41563a095bd70c78e2dda13d0f8cb922c71c90fffd0a0044c65173a66e83",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Size4-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Size4-Regular.woff",
		"3bc6ecaae7ecf6f8d7f8ea07bddbaca8601e2cb6e89d6aef58e78d9f6d8a398f",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Typewriter-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Typewriter-Regular.woff",
		"c56da8d69f1a0208b8e0703656c3264b6dd748bd452524c82b0385b60f6a68c1",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Vector-Bold.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Vector-Bold.woff",
		"36e0d72d8a7afc696a3e7a5c7369807634640de9b02257bca447bdf126221a27",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Vector-Regular.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Vector-Regular.woff",
		"72bc573386dd1d48c5bbd286302ca9e3400c5eb0f298e7cfbf945b9d08fc688f",
	)
	getAsset(
		"https://github.com/mathjax/MathJax/raw/refs/tags/3.2.2/es5/output/chtml/fonts/woff-v2/MathJax_Zero.woff",
		"./internal/server/static/output/chtml/fonts/woff-v2/MathJax_Zero.woff",
		"481e39042508ae313a60618af1e37146ab93e9324c98e4c78b8f17fe55d41e0b",
	)
}

func fatal[T any](v T, err error) T {
	if err != nil {
		log.Fatalln(err)
	}

	return v
}

func getAsset(url string, dest string, wantSum string) {
	// Download file
	resp := fatal(http.Get(url))
	defer resp.Body.Close()

	// Copy result to a buffer
	bs := fatal(io.ReadAll(resp.Body))

	// Check file checksum
	h := sha256.New()
	fatal(io.Copy(h, bytes.NewBuffer(bs)))
	gotSum := hex.EncodeToString(h.Sum(nil))
	if wantSum != gotSum {
		log.Fatalf(`Invalid sha256sum for %s:
Want: %s
Got: %s
`, url, wantSum, gotSum)
	}

	// Copy file to destination
	os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	out := fatal(os.Create(dest))
	defer out.Close()
	fatal(io.Copy(out, bytes.NewBuffer(bs)))

	fmt.Printf("Generated %s successfully\n", dest)
}

package regctl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/types/ref"
	"github.com/tidwall/gjson"
)

type AnonymousRegistry struct {}

func NewAnonymousRegistry() *AnonymousRegistry {
	return &AnonymousRegistry{
	}
}

func (rg *AnonymousRegistry) httpDo(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, "https://ghcr.io") {
		req.Header.Set("Authorization" , "Bearer QQ==")
	}
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err = http.DefaultClient.Do(req)
	return
}

func (rg *AnonymousRegistry) GetTags(ctx context.Context, refName string) (tags []string, err error) {
	r, err := ref.New(refName)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", r.Registry, r.Repository)
	resp, err := rg.httpDo(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	for _, x := range gjson.GetBytes(respBody, "tags").Array() {
		tags = append(tags, x.String())
	}
	return tags, nil
}

func (rg *AnonymousRegistry) GetManifest(ctx context.Context, refName string) (manifest string, err error) {
	r, err := ref.New(refName)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", r.Registry, r.Repository, r.Tag)
	resp, err := rg.httpDo(ctx, http.MethodGet, url, map[string]string{
		"Accept": "application/vnd.oci.image.index.v1+json",
	}, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

func (rg *AnonymousRegistry) DownloadBlob(ctx context.Context, refName string, sha256 string, outWriter io.Writer) error {
	r, err := ref.New(refName)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/v2/%s/blobs/sha256:%s", r.Registry, r.Repository, sha256)
	resp, err := rg.httpDo(ctx, http.MethodGet, url, map[string]string{
		"Accept": "application/vnd.oci.image.index.v1+json",
	}, nil)
	if err != nil {
		return  err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	_, err = io.Copy(outWriter, resp.Body)
	return err
}

type Registry struct {
	AnonymousRegistry
	reg string
	user string
	pass string
}

func NewRegistry(reg string, user string, pass string) *Registry {
	return &Registry{
		reg: reg,
		user: user,
		pass: pass,
	}
}

func (rg *Registry) getRegClient() *regclient.RegClient {
	host := config.HostNewName(rg.reg)
	host.User = rg.user
	host.Pass = rg.pass
	rc := regclient.New(regclient.WithConfigHost(*host))
	return rc
}

func (rg *Registry) Login() error {
	rc := rg.getRegClient()
	r, err := ref.NewHost(rg.reg)
	if err != nil {
		return err
	}
	_, err = rc.Ping(context.Background(), r)
	return err
}

func (rg *Registry) ImageCopy(ctx context.Context, ociRootDir string, imageRefWithoutHost string) error {
	srcRef := fmt.Sprintf("ocidir://%s", ociRootDir)
	dstRef := rg.GetRefFullName(imageRefWithoutHost)
	rSrc, err := ref.New(srcRef)
	if err != nil {
		return err
	}
	rTgt, err := ref.New(dstRef)
	if err != nil {
		return err
	}
	err = rg.getRegClient().ImageCopy(ctx, rSrc, rTgt, regclient.ImageWithReferrers())
	return err
}

func (rg *Registry) GetRefFullName(imageRefWithoutHost string) string {
	return fmt.Sprintf("%s/%s", rg.reg, imageRefWithoutHost)
}

func (rg *Registry) GetVersion(refName string) string {
	if strings.Contains(refName, ":") {
		ss := strings.Split(refName, ":")
		return ss[len(ss)-1]
	}
	return "latest"
}


package squadron

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/miracl/conflate"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func init() {
	// define the unmarshallers for the given file extensions, blank extension is the global unmarshaller
	conflate.Unmarshallers = conflate.UnmarshallerMap{
		".yaml": {conflate.YAMLUnmarshal},
		".yml":  {conflate.YAMLUnmarshal},
	}
}

type TemplateVars map[string]interface{}

func (tv *TemplateVars) add(name string, value interface{}) {
	(*tv)[name] = value
}

func executeFileTemplate(text string, templateVars interface{}, errorOnMissing bool) ([]byte, error) {
	templateFunctions := template.FuncMap{}
	templateFunctions["env"] = env
	templateFunctions["op"] = onePassword
	templateFunctions["base64"] = base64
	templateFunctions["default"] = defaultIndex
	templateFunctions["indent"] = indent
	templateFunctions["file"] = file
	templateFunctions["git"] = git

	tpl, err := template.New("squadron").Delims("<%", "%>").Funcs(templateFunctions).Parse(text)
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if errorOnMissing {
		tpl = tpl.Option("missingkey=error")
	}
	if err := tpl.Funcs(templateFunctions).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func env(name string) (string, error) {
	if value := os.Getenv(name); value == "" {
		return "", fmt.Errorf("env variable %q was empty", name)
	} else {
		return value, nil
	}
}

func file(v string) (string, error) {
	if v == "" {
		return "", nil
	} else if bs, err := ioutil.ReadFile(v); err != nil {
		return "", err
	} else {
		return string(bytes.TrimSpace(bs)), nil
	}
}

func defaultIndex(v map[string]interface{}, index string, def interface{}) interface{} {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}

func base64(v string) string {
	return b64.StdEncoding.EncodeToString([]byte(v))
}

func executeSquadronTemplate(text string, c *Configuration, tv TemplateVars) error {
	// execute without errors to get existing values
	out, err := executeFileTemplate(text, tv, false)
	if err != nil {
		return errors.Wrap(err, "failed to execute initial file template")
	}
	var vars map[string]interface{}
	if err := yaml.Unmarshal(out, &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	if value, ok := vars["global"]; ok {
		replace(value)
		tv.add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		replace(value)
		tv.add("Squadron", value)
	}
	out, err = executeFileTemplate(text, tv, true)
	if err != nil {
		return errors.Wrap(err, "failed to execute second file template")
	}
	if err := yaml.Unmarshal(out, &c); err != nil {
		return err
	}
	return nil
}

func replace(in interface{}) {
	if value, ok := in.(map[string]interface{}); ok {
		for k, v := range value {
			if strings.Contains(k, "-") {
				value[strings.Replace(k, "-", "_", -1)] = v
				delete(value, k)
			}
			replace(v)
		}
	}
}

func mergeSquadronFiles(files []string, c *Configuration, tv TemplateVars) error {
	// step 1: merge 'valid' yaml files
	mergedFiles, err := conflate.FromFiles(files...)
	if err != nil {
		return errors.Wrap(err, "failed to conflate files")
	}
	var data interface{}
	if err := mergedFiles.Unmarshal(&data); err != nil {
		return errors.Wrap(err, "failed to unmarshal data")
	}
	mergedBytes, err := mergedFiles.MarshalYAML()
	if err != nil {
		return errors.Wrap(err, "failed to marshal yaml")
	}

	// TODO print out YAML on debug

	// step 2: render template
	if err := executeSquadronTemplate(string(mergedBytes), c, tv); err != nil {
		return err
	}

	return nil
}

func git(action string) (string, error) {
	cmd := exec.Command("git")

	switch action {
	case "tag":
		cmd.Args = append(cmd.Args, "describe", "--tags", "--always")
	case "commitsha":
		cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD")
	case "abbrevcommitsha":
		cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD", "--abbrev-commit")
	}
	res, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(res)), nil
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return strings.Replace(v, "\n", "\n"+pad, -1)
}

func onePassword(account, uuid, field string) (string, error) {
	// validate command
	if _, err := exec.LookPath("op"); err != nil {
		fmt.Println("Your templates includes a call to 1Password, please install it:")
		fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
		return "", err
	}

	// validate session
	if os.Getenv(fmt.Sprintf("OP_SESSION_%s", account)) == "" {
		if err := onePasswordSignIn(account); err != nil {
			return "", err
		}
	}

	res, err := onePasswordGet(uuid, field)
	if err != nil && strings.Contains(res, "You are not currently signed in") {
		// retry with login
		if err := onePasswordSignIn(account); err != nil {
			return "", err
		} else if res, err = onePasswordGet(uuid, field); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return res, nil
}

func onePasswordGet(uuid, field string) (string, error) {
	res, err := exec.Command("op", "get", "item", uuid, "--fields", field).CombinedOutput()
	return string(res), err
}

func onePasswordSignIn(account string) error {
	fmt.Println("Your templates includes a call to 1Password, please sign to retrieve your session token:")

	// create command
	cmd := exec.Command("op", "signin", account, "--raw")

	// use multi writer to handle password prompt
	var stdoutBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stdin = os.Stdin

	// start the process and wait till it's finished
	if err := cmd.Start(); err != nil {
		return err
	} else if err := cmd.Wait(); err != nil {
		return err
	}

	if token := strings.TrimSuffix(stdoutBuf.String(), "\n"); token == "" {
		fmt.Println(fmt.Sprintf("Failed to login into your '%s' account! Please refer to the manual:", account))
		fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
		return errors.New("failed to retrieve 1password session token")
	} else if err := os.Setenv(fmt.Sprintf("OP_SESSION_%s", account), token); err != nil {
		return err
	} else {
		fmt.Println("NOTE: If you want to skip this step, run:")
		fmt.Println(fmt.Sprintf("export OP_SESSION_%s=%s", account, token))
	}

	return nil
}

# Argo Sidecar
This is a hack to be used for the issue - https://github.com/argoproj/argo-workflows/issues/12919

# How to use
1. Argo Templates
```yaml annotate
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: sidecar-
spec:
  entrypoint: sample-entrypoint
  templates:
    - name: template
      container:
        # Here is your main container - please define that as you need
      # Add the linkerd-killer sidecar, 
      sidecars:
        - name: linkerd-killer
          image: docker.io/athukorala/argo-sidecar-linkerd:latest
          command: [/bin/server, -podname, "{{pod.name}}", -namespace, "{{workflow.namespace}}"]
```

2. Golang SDK
```go annotate
[]wfv1.Template{
		{
			Name: workflowUtils.Entrypoint(w.job.Type),
			Steps: []wfv1.ParallelSteps{
				{
					Steps: []wfv1.WorkflowStep{
						{
							Name:     "hello-process",
							Template: "hello-template",
						},
					},
				},
			},
		},
		{
			Name: "hello-template",
			Inputs: wfv1.Inputs{
				Parameters: nil,
				Artifacts:  inputArtifacts,
			},
			Container: &corev1.Container{
				Image:           "",
				ImagePullPolicy: "",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "",
						MountPath: "",
					},
				},
				Resources:  "",
				WorkingDir: "",
				Command:    []string{"sh"},
				Args:       []string{"-c", ""},
			},
			Sidecars: []wfv1.UserContainer{
				{
					Container: corev1.Container{
						Name:    "linkerd-killer",
						Image:   "docker.io/athukorala/argo-sidecar-linkerd:latest",
						Command: []string{"bin/server"},
						Args:    []string{"-podname", "{{pod.name}}", "-namespace", "{{workflow.namespace}}"},
					},
				},
			},
			Outputs: wfv1.Outputs{
				Parameters: []wfv1.Parameter{},
				Artifacts: "",
			},
		},
	}
```
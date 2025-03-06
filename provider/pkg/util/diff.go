package util

import (
	"fmt"
	"slices"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

func StandardDiff(req *pulumirpc.DiffRequest, replaceProperties []string, anyDiffAsReplace bool) (*pulumirpc.DiffResponse, error) {
	olds, news, err := StandardOldNews(req)
	if err != nil {
		return nil, err
	}

	return StandardDiffNoMarshal(olds, news, replaceProperties, anyDiffAsReplace)
}

func StandardOldNews(req *pulumirpc.DiffRequest) (olds resource.PropertyMap, news resource.PropertyMap, err error) {
	olds, err = plugin.UnmarshalProperties(req.GetOldInputs(), StandardUnmarshal)
	if err != nil {
		return nil, nil, err
	}

	news, err = plugin.UnmarshalProperties(req.GetNews(), StandardUnmarshal)
	if err != nil {
		return nil, nil, err
	}

	return olds, news, nil
}

func StandardDiffNoMarshal(olds resource.PropertyMap, news resource.PropertyMap, replaceProperties []string, anyDiffAsReplace bool) (*pulumirpc.DiffResponse, error) {
	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)
	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	for k, v := range dd {
		if anyDiffAsReplace || slices.Contains(replaceProperties, k) {
			v.Kind = v.Kind.AsReplace()
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:             changes,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: true,
	}, nil
}

// For now, access tokens and webhooks cannot be fully standardized due to need to support __inputs
// PSPV1 will remove this deprecated way to store inputs and then these methods will be removed
func DeprecatedInputDiff(olds resource.PropertyMap, news resource.PropertyMap, replaceProperties []string, anyDiffAsReplace bool) (*pulumirpc.DiffResponse, error) {
	inputs, ok := olds["__inputs"]
	if !ok {
		return nil, fmt.Errorf("missing __inputs property")
	}

	return StandardDiffNoMarshal(inputs.ObjectValue(), news, replaceProperties, anyDiffAsReplace)
}

func DeprecatedOptionalInputDiff(olds resource.PropertyMap, news resource.PropertyMap, replaceProperties []string, anyDiffAsReplace bool) (*pulumirpc.DiffResponse, error) {
	if oldInputs, ok := olds["__inputs"]; ok && oldInputs.IsObject() {
		for k, v := range oldInputs.ObjectValue() {
			olds[k] = v
		}
	}

	return StandardDiffNoMarshal(olds, news, replaceProperties, anyDiffAsReplace)
}

func DeprecatedOldNews(req *pulumirpc.DiffRequest) (olds resource.PropertyMap, news resource.PropertyMap, err error) {
	olds, err = plugin.UnmarshalProperties(req.GetOlds(), StandardUnmarshal)
	if err != nil {
		return nil, nil, err
	}

	news, err = plugin.UnmarshalProperties(req.GetNews(), StandardUnmarshal)
	if err != nil {
		return nil, nil, err
	}

	return olds, news, nil
}

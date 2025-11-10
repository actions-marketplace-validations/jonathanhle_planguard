package functions

import (
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/jonathanhle/planguard/pkg/parser"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// BuildFunctions creates the complete function map for evaluation
func BuildFunctions(ctx *parser.ScanContext) map[string]function.Function {
	// Start with go-cty's standard library functions
	functions := make(map[string]function.Function)

	// String functions
	functions["upper"] = stdlib.UpperFunc
	functions["lower"] = stdlib.LowerFunc
	functions["trim"] = stdlib.TrimFunc
	functions["trimspace"] = stdlib.TrimSpaceFunc
	functions["trimprefix"] = stdlib.TrimPrefixFunc
	functions["trimsuffix"] = stdlib.TrimSuffixFunc
	functions["split"] = stdlib.SplitFunc
	functions["join"] = stdlib.JoinFunc
	functions["replace"] = stdlib.ReplaceFunc
	functions["substr"] = stdlib.SubstrFunc
	functions["format"] = stdlib.FormatFunc
	functions["formatlist"] = stdlib.FormatListFunc
	functions["indent"] = stdlib.IndentFunc
	functions["chomp"] = stdlib.ChompFunc
	functions["regex"] = stdlib.RegexFunc
	functions["regexall"] = stdlib.RegexAllFunc

	// Collection functions
	functions["length"] = stdlib.LengthFunc
	functions["concat"] = stdlib.ConcatFunc
	functions["flatten"] = stdlib.FlattenFunc
	functions["contains"] = stdlib.ContainsFunc
	functions["distinct"] = stdlib.DistinctFunc
	functions["keys"] = stdlib.KeysFunc
	functions["values"] = stdlib.ValuesFunc
	functions["lookup"] = stdlib.LookupFunc
	functions["element"] = stdlib.ElementFunc
	functions["index"] = stdlib.IndexFunc
	functions["reverse"] = stdlib.ReverseFunc
	functions["slice"] = stdlib.SliceFunc
	functions["sort"] = stdlib.SortFunc
	functions["chunklist"] = stdlib.ChunklistFunc
	functions["compact"] = stdlib.CompactFunc
	functions["coalesce"] = stdlib.CoalesceFunc
	functions["coalescelist"] = stdlib.CoalesceListFunc
	functions["merge"] = stdlib.MergeFunc
	functions["zipmap"] = stdlib.ZipmapFunc

	// Numeric functions
	functions["min"] = stdlib.MinFunc
	functions["max"] = stdlib.MaxFunc
	functions["ceil"] = stdlib.CeilFunc
	functions["floor"] = stdlib.FloorFunc
	functions["abs"] = stdlib.AbsoluteFunc
	functions["parseint"] = stdlib.ParseIntFunc
	functions["int"] = stdlib.IntFunc

	// Type conversion functions
	functions["tonumber"] = stdlib.IntFunc // Parse to number

	// Add HCL extension functions
	functions["try"] = tryfunc.TryFunc
	functions["can"] = tryfunc.CanFunc
	functions["hasindex"] = stdlib.HasIndexFunc

	// Add YAML functions
	functions["yamldecode"] = ctyyaml.YAMLDecodeFunc
	functions["yamlencode"] = ctyyaml.YAMLEncodeFunc

	// Add custom encoding functions
	functions["jsondecode"] = JSONDecodeFunc
	functions["jsonencode"] = JSONEncodeFunc
	functions["base64encode"] = Base64EncodeFunc
	functions["base64decode"] = Base64DecodeFunc
	functions["base64gzip"] = Base64GzipFunc
	functions["urlencode"] = URLEncodeFunc

	// Add custom crypto functions
	functions["md5"] = MD5Func
	functions["sha1"] = SHA1Func
	functions["sha256"] = SHA256Func
	functions["sha512"] = SHA512Func
	functions["base64sha256"] = Base64SHA256Func
	functions["base64sha512"] = Base64SHA512Func
	functions["bcrypt"] = BcryptFunc
	functions["uuid"] = UUIDFunc
	functions["uuidv5"] = UUIDV5Func

	// Add CIDR functions
	functions["cidrhost"] = CIDRHostFunc
	functions["cidrnetmask"] = CIDRNetmaskFunc
	functions["cidrsubnet"] = CIDRSubnetFunc
	functions["cidrsubnets"] = CIDRSubnetsFunc

	// Add datetime functions
	functions["timestamp"] = TimestampFunc
	functions["formatdate"] = FormatDateFunc
	functions["timeadd"] = TimeAddFunc

	// Add domain-specific functions
	functions["resources"] = ResourcesFunc(ctx)
	functions["resources_in_file"] = ResourcesInFileFunc(ctx)
	functions["day_of_week"] = DayOfWeekFunc
	functions["git_branch"] = GitBranchFunc
	functions["has"] = HasFunc
	functions["anytrue"] = AnyTrueFunc
	functions["alltrue"] = AllTrueFunc

	// Add security functions
	functions["contains_function_call"] = ContainsFunctionCallFunc(ctx)

	// Add utility functions
	functions["glob_match"] = GlobMatchFunc
	functions["regex_match"] = RegexMatchFunc

	return functions
}

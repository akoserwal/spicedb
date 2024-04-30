package development

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	developerv1 "github.com/authzed/spicedb/pkg/proto/developer/v1"
)

func TestWarnings(t *testing.T) {
	tcs := []struct {
		name            string
		schema          string
		expectedWarning *developerv1.DeveloperWarning
	}{
		{
			name: "no warnings",
			schema: `definition user {}
			
			definition group {
				relation direct_member: user
				permission member = direct_member
			}

			definition document {
				relation viewer: user | group#member
				permission view = viewer
			}
			`,
		},
		{
			name: "permission referencing itself",
			schema: `definition test {
				permission view = view	
			}`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Permission \"view\" references itself, which will cause an error to be raised due to infinite recursion",
				Line:    2,
				Column:  5,
			},
		},
		{
			name: "permission referencing itself, nested",
			schema: `definition test {
				relation viewer: test
				relation editor: test
				permission view = viewer + (editor & view)	
			}`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Permission \"view\" references itself, which will cause an error to be raised due to infinite recursion",
				Line:    4,
				Column:  5,
			},
		},
		{
			name: "arrow referencing relation",
			schema: `definition group {
				relation member: user
			}
			
			definition user {}

			definition document {
				relation group: group
				permission view = group->member
			}
			`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Arrow `group->member` under permission \"view\" references relation \"member\" on definition \"group\"; it is recommended to point to a permission",
				Line:    9,
				Column:  5,
			},
		},
		{
			name: "arrow referencing unknown relation",
			schema: `definition group {
			}
			
			definition user {}

			definition document {
				relation group: group
				permission view = group->member
			}
			`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Arrow `group->member` under permission \"view\" references relation/permission \"member\" that does not exist on any subject types of relation \"group\"",
				Line:    8,
				Column:  5,
			},
		},
		{
			name: "arrow referencing subject relation",
			schema: `definition group {
				relation direct_member: user
				permission member = direct_member
			}
			
			definition user {}

			definition document {
				relation parent_group: group#member
				permission view = parent_group->member
			}
			`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Arrow `parent_group->member` under permission \"view\" references relation \"parent_group\" that has relation \"member\" on subject \"group\": *the subject relation will be ignored for the arrow*",
				Line:    10,
				Column:  5,
			},
		},
		{
			name: "relation referencing its parent definition in its name",
			schema: `definition user {}

			definition document {
				relation viewer: user
				permission view_document = viewer
			}`,
			expectedWarning: &developerv1.DeveloperWarning{
				Message: "Permission \"view_document\" references parent type \"document\" in its name; it is recommended to drop the suffix",
				Line:    5,
				Column:  5,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			devCtx, devErr, err := NewDevContext(context.Background(), &developerv1.RequestContext{
				Schema: tc.schema,
			})
			require.NoError(t, err)
			require.Empty(t, devErr)

			warnings, err := GetWarnings(context.Background(), devCtx)
			require.NoError(t, err)

			if tc.expectedWarning == nil {
				require.Empty(t, warnings)
			} else {
				require.Len(t, warnings, 1, "expected exactly one warning")
				require.Equal(t, tc.expectedWarning, warnings[0])
			}
		})
	}
}

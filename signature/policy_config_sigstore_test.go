package signature

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// xNewPRSigstoreSigned is like NewPRSigstoreSigned, except it must not fail.
func xNewPRSigstoreSigned(options ...PRSigstoreSignedOption) PolicyRequirement {
	pr, err := NewPRSigstoreSigned(options...)
	if err != nil {
		panic("xNewPRSigstoreSigned failed")
	}
	return pr
}

func TestNewPRSigstoreSigned(t *testing.T) {
	const testKeyPath = "/foo/bar"
	const testKeyPath2 = "/baz/bar"
	testKeyData := []byte("abc")
	testKeyData2 := []byte("def")
	testFulcio, err := NewPRSigstoreSignedFulcio(
		PRSigstoreSignedFulcioWithCAPath("fixtures/fulcio_v1.crt.pem"),
		PRSigstoreSignedFulcioWithOIDCIssuer("https://github.com/login/oauth"),
		PRSigstoreSignedFulcioWithSubjectEmail("mitr@redhat.com"),
	)
	require.NoError(t, err)
	const testRekorKeyPath = "/foo/baz"
	testRekorKeyData := []byte("def")
	testIdentity := NewPRMMatchRepoDigestOrExact()

	// Success: combinatoric combinations of key source and Rekor uses
	for _, c := range []struct {
		options       []PRSigstoreSignedOption
		requiresRekor bool
		expected      prSigstoreSigned
	}{
		{
			options: []PRSigstoreSignedOption{
				PRSigstoreSignedWithKeyPath(testKeyPath),
				PRSigstoreSignedWithSignedIdentity(testIdentity),
			},
			expected: prSigstoreSigned{
				prCommon:       prCommon{prTypeSigstoreSigned},
				KeyPath:        testKeyPath,
				KeyPaths:       nil,
				KeyData:        nil,
				KeyDatas:       nil,
				Fulcio:         nil,
				SignedIdentity: testIdentity,
			},
		},
		{
			options: []PRSigstoreSignedOption{
				PRSigstoreSignedWithKeyPaths([]string{testKeyPath, testKeyPath2}),
				PRSigstoreSignedWithSignedIdentity(testIdentity),
			},
			expected: prSigstoreSigned{
				prCommon:       prCommon{prTypeSigstoreSigned},
				KeyPath:        "",
				KeyPaths:       []string{testKeyPath, testKeyPath2},
				KeyData:        nil,
				KeyDatas:       nil,
				Fulcio:         nil,
				SignedIdentity: testIdentity,
			},
		},
		{
			options: []PRSigstoreSignedOption{
				PRSigstoreSignedWithKeyData(testKeyData),
				PRSigstoreSignedWithSignedIdentity(testIdentity),
			},
			expected: prSigstoreSigned{
				prCommon:       prCommon{prTypeSigstoreSigned},
				KeyPath:        "",
				KeyPaths:       nil,
				KeyData:        testKeyData,
				KeyDatas:       nil,
				Fulcio:         nil,
				SignedIdentity: testIdentity,
			},
		},
		{
			options: []PRSigstoreSignedOption{
				PRSigstoreSignedWithKeyDatas([][]byte{testKeyData, testKeyData2}),
				PRSigstoreSignedWithSignedIdentity(testIdentity),
			},
			expected: prSigstoreSigned{
				prCommon:       prCommon{prTypeSigstoreSigned},
				KeyPath:        "",
				KeyPaths:       nil,
				KeyData:        nil,
				KeyDatas:       [][]byte{testKeyData, testKeyData2},
				Fulcio:         nil,
				SignedIdentity: testIdentity,
			},
		},
		{
			options: []PRSigstoreSignedOption{
				PRSigstoreSignedWithFulcio(testFulcio),
				PRSigstoreSignedWithSignedIdentity(testIdentity),
			},
			requiresRekor: true,
			expected: prSigstoreSigned{
				prCommon:       prCommon{prTypeSigstoreSigned},
				KeyPath:        "",
				KeyPaths:       nil,
				KeyData:        nil,
				KeyDatas:       nil,
				Fulcio:         testFulcio,
				SignedIdentity: testIdentity,
			},
		},
	} {
		for _, c2 := range []struct {
			rekorOptions  []PRSigstoreSignedOption
			rekorExpected prSigstoreSigned
		}{
			{ // No Rekor
				rekorOptions:  []PRSigstoreSignedOption{},
				rekorExpected: prSigstoreSigned{},
			},
			{
				rekorOptions: []PRSigstoreSignedOption{
					PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
				},
				rekorExpected: prSigstoreSigned{
					RekorPublicKeyPath: testRekorKeyPath,
				},
			},
			{
				rekorOptions: []PRSigstoreSignedOption{
					PRSigstoreSignedWithRekorPublicKeyPaths([]string{testRekorKeyPath, testKeyPath}),
				},
				rekorExpected: prSigstoreSigned{
					RekorPublicKeyPaths: []string{testRekorKeyPath, testKeyPath},
				},
			},
			{
				rekorOptions: []PRSigstoreSignedOption{
					PRSigstoreSignedWithRekorPublicKeyData(testRekorKeyData),
				},
				rekorExpected: prSigstoreSigned{
					RekorPublicKeyData: testRekorKeyData,
				},
			},
			{
				rekorOptions: []PRSigstoreSignedOption{
					PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{testRekorKeyData, testKeyData}),
				},
				rekorExpected: prSigstoreSigned{
					RekorPublicKeyDatas: [][]byte{testRekorKeyData, testKeyData},
				},
			},
		} {
			// Special-case this rejected combination:
			if c.requiresRekor && len(c2.rekorOptions) == 0 {
				continue
			}
			pr, err := newPRSigstoreSigned(append(c.options, c2.rekorOptions...)...)
			require.NoError(t, err)
			expected := c.expected // A shallow copy
			expected.RekorPublicKeyPath = c2.rekorExpected.RekorPublicKeyPath
			expected.RekorPublicKeyPaths = c2.rekorExpected.RekorPublicKeyPaths
			expected.RekorPublicKeyData = c2.rekorExpected.RekorPublicKeyData
			expected.RekorPublicKeyDatas = c2.rekorExpected.RekorPublicKeyDatas
			assert.Equal(t, &expected, pr)
		}
	}

	testFulcio2, err := NewPRSigstoreSignedFulcio(
		PRSigstoreSignedFulcioWithCAPath("fixtures/fulcio_v1.crt.pem"),
		PRSigstoreSignedFulcioWithOIDCIssuer("https://github.com/login/oauth"),
		PRSigstoreSignedFulcioWithSubjectEmail("test-user@example.com"),
	)
	require.NoError(t, err)
	for _, c := range [][]PRSigstoreSignedOption{
		{}, // None of keyPath, keyPaths, keyData, keyDatas, fulcio specified
		{ // Both keyPath and keyData specified
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithKeyData(testKeyData),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Both keyPath and fulcio specified
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithFulcio(testFulcio),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Both keyData and fulcio specified
			PRSigstoreSignedWithKeyData(testKeyData),
			PRSigstoreSignedWithFulcio(testFulcio),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate keyPath
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithKeyPath(testKeyPath + "1"),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Empty keypaths
			PRSigstoreSignedWithKeyPaths([]string{}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate keyPaths
			PRSigstoreSignedWithKeyPaths([]string{testKeyPath, testKeyPath2}),
			PRSigstoreSignedWithKeyPaths([]string{testKeyPath + "1", testKeyPath2 + "1"}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // keyPath & keyPaths both set
			PRSigstoreSignedWithKeyPath("foobar"),
			PRSigstoreSignedWithKeyPaths([]string{"foobar"}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate keyData
			PRSigstoreSignedWithKeyData(testKeyData),
			PRSigstoreSignedWithKeyData([]byte("def")),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Empty keyDatas
			PRSigstoreSignedWithKeyDatas([][]byte{}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate keyDatas
			PRSigstoreSignedWithKeyDatas([][]byte{testKeyData, testKeyData2}),
			PRSigstoreSignedWithKeyDatas([][]byte{append(testKeyData, 'a'), append(testKeyData2, 'a')}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // keyData & keyDatas both set
			PRSigstoreSignedWithKeyData([]byte("bar")),
			PRSigstoreSignedWithKeyDatas([][]byte{[]byte("foo")}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate fulcio
			PRSigstoreSignedWithFulcio(testFulcio),
			PRSigstoreSignedWithFulcio(testFulcio2),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // fulcio without Rekor
			PRSigstoreSignedWithFulcio(testFulcio),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Both rekorKeyPath and rekorKeyData specified
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithRekorPublicKeyData(testRekorKeyData),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate rekorKeyPath
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath + "1"),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Both rekorKeyPath and rekorKeyPaths specified
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPath(testRekorKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPaths([]string{testRekorKeyPath, testKeyPath}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Empty rekorKeyPaths
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPaths([]string{}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate rekorKeyPaths
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyPaths([]string{testRekorKeyPath, testKeyPath}),
			PRSigstoreSignedWithRekorPublicKeyPaths([]string{testRekorKeyPath + "1", testKeyPath + "1"}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate rekorKeyData
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyData(testRekorKeyData),
			PRSigstoreSignedWithRekorPublicKeyData([]byte("def")),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Both rekorKeyData and rekorKeyDatas specified
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyData(testRekorKeyData),
			PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{testRekorKeyData, []byte("def")}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Empty rekorKeyDatas
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Duplicate rekorKeyData
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{testRekorKeyData, testKeyData}),
			PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{[]byte("abc"), []byte("def")}),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
		},
		{ // Missing signedIdentity
			PRSigstoreSignedWithKeyPath(testKeyPath),
		},
		{ // Duplicate signedIdentity
			PRSigstoreSignedWithKeyPath(testKeyPath),
			PRSigstoreSignedWithSignedIdentity(testIdentity),
			PRSigstoreSignedWithSignedIdentity(newPRMMatchRepository()),
		},
	} {
		_, err = newPRSigstoreSigned(c...)
		assert.Error(t, err)
	}
}

func TestNewPRSigstoreSignedKeyPath(t *testing.T) {
	const testPath = "/foo/bar"
	signedIdentity := NewPRMMatchRepoDigestOrExact()
	_pr, err := NewPRSigstoreSignedKeyPath(testPath, signedIdentity)
	require.NoError(t, err)
	pr, ok := _pr.(*prSigstoreSigned)
	require.True(t, ok)
	assert.Equal(t, &prSigstoreSigned{
		prCommon:       prCommon{Type: prTypeSigstoreSigned},
		KeyPath:        testPath,
		SignedIdentity: NewPRMMatchRepoDigestOrExact(),
	}, pr)
}

func TestNewPRSigstoreSignedKeyData(t *testing.T) {
	testData := []byte("abc")
	signedIdentity := NewPRMMatchRepoDigestOrExact()
	_pr, err := NewPRSigstoreSignedKeyData(testData, signedIdentity)
	require.NoError(t, err)
	pr, ok := _pr.(*prSigstoreSigned)
	require.True(t, ok)
	assert.Equal(t, &prSigstoreSigned{
		prCommon:       prCommon{Type: prTypeSigstoreSigned},
		KeyData:        testData,
		SignedIdentity: NewPRMMatchRepoDigestOrExact(),
	}, pr)
}

// Return the result of modifying validJSON with fn and unmarshaling it into *pr
func tryUnmarshalModifiedSigstoreSigned(t *testing.T, pr *prSigstoreSigned, validJSON []byte, modifyFn func(mSA)) error {
	var tmp mSA
	err := json.Unmarshal(validJSON, &tmp)
	require.NoError(t, err)

	modifyFn(tmp)

	*pr = prSigstoreSigned{}
	return jsonUnmarshalFromObject(t, tmp, &pr)
}

func TestPRSigstoreSignedUnmarshalJSON(t *testing.T) {
	keyDataTests := policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSignedKeyData([]byte("abc"), NewPRMMatchRepoDigestOrExact())
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		breakFns: []func(mSA){
			// The "type" field is missing
			func(v mSA) { delete(v, "type") },
			// Wrong "type" field
			func(v mSA) { v["type"] = 1 },
			func(v mSA) { v["type"] = "this is invalid" },
			// Extra top-level sub-object
			func(v mSA) { v["unexpected"] = 1 },
			// All of "keyPath", "keyPaths", "keyData", "keyDatas", and "fulcio" is missing
			func(v mSA) { delete(v, "keyData") },
			// Both "keyPath" and "keyData" is present
			func(v mSA) { v["keyPath"] = "/foo/bar" },
			// Both "keyPaths" and "keyData" is present
			func(v mSA) { v["keyPaths"] = []string{"/foo/bar", "/foo/baz"} },
			// Both "keyData" and "keyDatas" is present
			func(v mSA) { v["keyDatas"] = [][]byte{[]byte("abc"), []byte("def")} },
			// Both "keyData" and "fulcio" is present
			func(v mSA) {
				v["fulcio"] = mSA{
					"caPath":       "/foo/baz",
					"oidcIssuer":   "https://example.com",
					"subjectEmail": "test@example.com",
				}
			},
			// Invalid "keyPath" field
			func(v mSA) { delete(v, "keyData"); v["keyPath"] = 1 },
			// Invalid "keyPaths" field
			func(v mSA) { delete(v, "keyData"); v["keyPaths"] = 1 },
			func(v mSA) { delete(v, "keyData"); v["keyPaths"] = mSA{} },
			func(v mSA) { delete(v, "keyData"); v["keyPaths"] = []string{} },
			// Invalid "keyData" field
			func(v mSA) { v["keyData"] = 1 },
			func(v mSA) { v["keyData"] = "this is invalid base64" },
			// Invalid "keyDatas" field
			func(v mSA) { delete(v, "keyData"); v["keyDatas"] = 1 },
			func(v mSA) { delete(v, "keyData"); v["keyDatas"] = mSA{} },
			func(v mSA) { delete(v, "keyData"); v["keyDatas"] = [][]byte{} },
			// Invalid "fulcio" field
			func(v mSA) { delete(v, "keyData"); v["fulcio"] = 1 },
			func(v mSA) { delete(v, "keyData"); v["fulcio"] = mSA{} },
			// "fulcio" is explicit nil
			func(v mSA) { delete(v, "keyData"); v["fulcio"] = nil },
			// Both "rekorKeyPath" and "rekorKeyData" is present
			func(v mSA) {
				v["rekorPublicKeyPath"] = "/foo/baz"
				v["rekorPublicKeyData"] = ""
			},
			// Invalid "rekorPublicKeyPath" field
			func(v mSA) { v["rekorPublicKeyPath"] = 1 },
			// Both "rekorKeyPath" and "rekorKeyPaths" is present
			func(v mSA) {
				v["rekorPublicKeyPath"] = "/foo/baz"
				v["rekorPublicKeyPaths"] = []string{"/baz/a", "/baz/b"}
			},
			// Invalid "rekorPublicKeyPaths" field
			func(v mSA) { v["rekorPublicKeyPaths"] = 1 },
			func(v mSA) { v["rekorPublicKeyPaths"] = mSA{} },
			func(v mSA) { v["rekorPublicKeyPaths"] = []string{} },
			// Invalid "rekorPublicKeyData" field
			func(v mSA) { v["rekorPublicKeyData"] = 1 },
			func(v mSA) { v["rekorPublicKeyData"] = "this is invalid base64" },
			// Both "rekorPublicKeyData" and "rekorPublicKeyDatas" is present
			func(v mSA) {
				v["rekorPublicKeyData"] = []byte("a")
				v["rekorPublicKeyDatas"] = [][]byte{[]byte("a"), []byte("b")}
			},
			// Invalid "rekorPublicKeyDatas" field
			func(v mSA) { v["rekorPublicKeyDatas"] = 1 },
			func(v mSA) { v["rekorPublicKeyDatas"] = mSA{} },
			func(v mSA) { v["rekorPublicKeyDatas"] = [][]byte{} },
			// Invalid "signedIdentity" field
			func(v mSA) { v["signedIdentity"] = "this is invalid" },
			// "signedIdentity" an explicit nil
			func(v mSA) { v["signedIdentity"] = nil },
		},
		duplicateFields: []string{"type", "keyData", "signedIdentity"},
	}
	keyDataTests.run(t)
	// Test keyPath and keyPath-specific duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSignedKeyPath("/foo/bar", NewPRMMatchRepoDigestOrExact())
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyPath", "signedIdentity"},
	}.run(t)
	// Test keyPaths and keyPaths-specific duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithKeyPaths([]string{"/foo/bar", "/foo/baz"}),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyPaths", "signedIdentity"},
	}.run(t)
	// Test keyDatas and keyDatas-specific duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithKeyDatas([][]byte{[]byte("abc"), []byte("def")}),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyDatas", "signedIdentity"},
	}.run(t)
	// Test Fulcio and rekorPublicKeyPath duplicate fields
	testFulcio, err := NewPRSigstoreSignedFulcio(
		PRSigstoreSignedFulcioWithCAPath("fixtures/fulcio_v1.crt.pem"),
		PRSigstoreSignedFulcioWithOIDCIssuer("https://github.com/login/oauth"),
		PRSigstoreSignedFulcioWithSubjectEmail("mitr@redhat.com"),
	)
	require.NoError(t, err)
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithFulcio(testFulcio),
				PRSigstoreSignedWithRekorPublicKeyPath("/foo/rekor"),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "fulcio", "rekorPublicKeyPath", "signedIdentity"},
	}.run(t)
	// Test rekorPublicKeyPaths duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithKeyPath("/foo/bar"),
				PRSigstoreSignedWithRekorPublicKeyPaths([]string{"/baz/a", "/baz/b"}),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyPath", "rekorPublicKeyPaths", "signedIdentity"},
	}.run(t)
	// Test rekorPublicKeyData duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithKeyPath("/foo/bar"),
				PRSigstoreSignedWithRekorPublicKeyData([]byte("foo")),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyPath", "rekorPublicKeyData", "signedIdentity"},
	}.run(t)
	// Test rekorPublicKeyDatas duplicate fields
	policyJSONUmarshallerTests[PolicyRequirement]{
		newDest: func() json.Unmarshaler { return &prSigstoreSigned{} },
		newValidObject: func() (PolicyRequirement, error) {
			return NewPRSigstoreSigned(
				PRSigstoreSignedWithKeyPath("/foo/bar"),
				PRSigstoreSignedWithRekorPublicKeyDatas([][]byte{[]byte("foo"), []byte("bar")}),
				PRSigstoreSignedWithSignedIdentity(NewPRMMatchRepoDigestOrExact()),
			)
		},
		otherJSONParser: newPolicyRequirementFromJSON,
		duplicateFields: []string{"type", "keyPath", "rekorPublicKeyDatas", "signedIdentity"},
	}.run(t)

	var pr prSigstoreSigned

	// Start with a valid JSON.
	_, validJSON := keyDataTests.validObjectAndJSON(t)

	// Various allowed modifications to the requirement
	allowedModificationFns := []func(mSA){
		// Delete the signedIdentity field
		func(v mSA) { delete(v, "signedIdentity") },
	}
	for _, fn := range allowedModificationFns {
		err := tryUnmarshalModifiedSigstoreSigned(t, &pr, validJSON, fn)
		require.NoError(t, err)
	}

	// Various ways to set signedIdentity to the default value
	signedIdentityDefaultFns := []func(mSA){
		// Set signedIdentity to the default explicitly
		func(v mSA) { v["signedIdentity"] = NewPRMMatchRepoDigestOrExact() },
		// Delete the signedIdentity field
		func(v mSA) { delete(v, "signedIdentity") },
	}
	for _, fn := range signedIdentityDefaultFns {
		err := tryUnmarshalModifiedSigstoreSigned(t, &pr, validJSON, fn)
		require.NoError(t, err)
		assert.Equal(t, NewPRMMatchRepoDigestOrExact(), pr.SignedIdentity)
	}
}

func TestNewPRSigstoreSignedFulcio(t *testing.T) {
	const testCAPath = "/foo/bar"
	testCAData := []byte("abc")
	const testOIDCIssuer = "https://example.com"
	const testSubjectEmail = "test@example.com"

	// Success:
	for _, c := range []struct {
		options  []PRSigstoreSignedFulcioOption
		expected prSigstoreSignedFulcio
	}{
		{
			options: []PRSigstoreSignedFulcioOption{
				PRSigstoreSignedFulcioWithCAPath(testCAPath),
				PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
				PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
			},
			expected: prSigstoreSignedFulcio{
				CAPath:       testCAPath,
				OIDCIssuer:   testOIDCIssuer,
				SubjectEmail: testSubjectEmail,
			},
		},
		{
			options: []PRSigstoreSignedFulcioOption{
				PRSigstoreSignedFulcioWithCAData(testCAData),
				PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
				PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
			},
			expected: prSigstoreSignedFulcio{
				CAData:       testCAData,
				OIDCIssuer:   testOIDCIssuer,
				SubjectEmail: testSubjectEmail,
			},
		},
	} {
		pr, err := newPRSigstoreSignedFulcio(c.options...)
		require.NoError(t, err)
		assert.Equal(t, &c.expected, pr)
	}

	for _, c := range [][]PRSigstoreSignedFulcioOption{
		{ // Neither caPath nor caData specified
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Both caPath and caData specified
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithCAData(testCAData),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Duplicate caPath
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithCAPath(testCAPath + "1"),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Duplicate caData
			PRSigstoreSignedFulcioWithCAData(testCAData),
			PRSigstoreSignedFulcioWithCAData([]byte("def")),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Missing oidcIssuer
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Duplicate oidcIssuer
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer + "1"),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
		},
		{ // Missing subjectEmail
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
		},
		{ // Duplicate subjectEmail
			PRSigstoreSignedFulcioWithCAPath(testCAPath),
			PRSigstoreSignedFulcioWithOIDCIssuer(testOIDCIssuer),
			PRSigstoreSignedFulcioWithSubjectEmail(testSubjectEmail),
			PRSigstoreSignedFulcioWithSubjectEmail("1" + testSubjectEmail),
		},
	} {
		_, err := newPRSigstoreSignedFulcio(c...)
		logrus.Errorf("%#v", err)
		assert.Error(t, err)
	}
}

func TestPRSigstoreSignedFulcioUnmarshalJSON(t *testing.T) {
	policyJSONUmarshallerTests[PRSigstoreSignedFulcio]{
		newDest: func() json.Unmarshaler { return &prSigstoreSignedFulcio{} },
		newValidObject: func() (PRSigstoreSignedFulcio, error) {
			return NewPRSigstoreSignedFulcio(
				PRSigstoreSignedFulcioWithCAPath("fixtures/fulcio_v1.crt.pem"),
				PRSigstoreSignedFulcioWithOIDCIssuer("https://github.com/login/oauth"),
				PRSigstoreSignedFulcioWithSubjectEmail("mitr@redhat.com"),
			)
		},
		otherJSONParser: nil,
		breakFns: []func(mSA){
			// Extra top-level sub-object
			func(v mSA) { v["unexpected"] = 1 },
			// Both of "caPath" and "caData" are missing
			func(v mSA) { delete(v, "caPath") },
			// Both "caPath" and "caData" is present
			func(v mSA) { v["caData"] = "" },
			// Invalid "caPath" field
			func(v mSA) { v["caPath"] = 1 },
			// Invalid "oidcIssuer" field
			func(v mSA) { v["oidcIssuer"] = 1 },
			// "oidcIssuer" is missing
			func(v mSA) { delete(v, "oidcIssuer") },
			// Invalid "subjectEmail" field
			func(v mSA) { v["subjectEmail"] = 1 },
			// "subjectEmail" is missing
			func(v mSA) { delete(v, "subjectEmail") },
		},
		duplicateFields: []string{"caPath", "oidcIssuer", "subjectEmail"},
	}.run(t)
	// Test caData specifics
	policyJSONUmarshallerTests[PRSigstoreSignedFulcio]{
		newDest: func() json.Unmarshaler { return &prSigstoreSignedFulcio{} },
		newValidObject: func() (PRSigstoreSignedFulcio, error) {
			return NewPRSigstoreSignedFulcio(
				PRSigstoreSignedFulcioWithCAData([]byte("abc")),
				PRSigstoreSignedFulcioWithOIDCIssuer("https://github.com/login/oauth"),
				PRSigstoreSignedFulcioWithSubjectEmail("mitr@redhat.com"),
			)
		},
		otherJSONParser: nil,
		breakFns: []func(mSA){
			// Invalid "caData" field
			func(v mSA) { v["caData"] = 1 },
			func(v mSA) { v["caData"] = "this is invalid base64" },
		},
		duplicateFields: []string{"caData", "oidcIssuer", "subjectEmail"},
	}.run(t)
}
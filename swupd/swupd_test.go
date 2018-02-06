package swupd

import "testing"

// Imported from swupd-server/test/functional/include-version-bump.
func TestIncludeVersionBump(t *testing.T) {
	ts := newTestSwupd(t, "include-version-bump-")
	defer ts.cleanup()

	// Version 10.
	ts.Bundles = []string{"test-bundle"}
	ts.write("image/10/test-bundle/foo", "foo")
	ts.createManifests(10)
	ts.createFullfiles(10)
	ts.createPack("os-core", 0, 10, ts.path("image"))
	ts.createPack("test-bundle", 0, 10, ts.path("image"))

	mustValidateZeroPack(t, ts.path("www/10/Manifest.os-core"), ts.path("www/10/pack-os-core-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/10/Manifest.test-bundle"), ts.path("www/10/pack-test-bundle-from-0.tar"))

	m10 := ts.parseManifest(10, "test-bundle")
	checkIncludes(t, m10, "os-core")
	checkFileInManifest(t, m10, 10, "/foo")

	// Version 20. Add "included" and "included-two" to be included by "test-bundle".
	ts.Bundles = append(ts.Bundles, "included", "included-two")
	ts.copyChroots(10, 20)
	ts.write("image/20/included/bar", "bar")
	ts.write("image/20/included-two/baz", "baz")
	ts.write("image/20/noship/test-bundle-includes", "included\nincluded-two")
	ts.createManifests(20)
	ts.createFullfiles(20)

	ts.createPack("os-core", 0, 20, ts.path("image"))
	ts.createPack("test-bundle", 0, 20, ts.path("image"))
	ts.createPack("included", 0, 20, ts.path("image"))
	ts.createPack("included-two", 0, 20, ts.path("image"))

	mustValidateZeroPack(t, ts.path("www/20/Manifest.os-core"), ts.path("www/20/pack-os-core-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/20/Manifest.test-bundle"), ts.path("www/20/pack-test-bundle-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/20/Manifest.included"), ts.path("www/20/pack-included-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/20/Manifest.included-two"), ts.path("www/20/pack-included-two-from-0.tar"))

	m20 := ts.parseManifest(20, "test-bundle")
	checkIncludes(t, m20, "os-core", "included", "included-two")

	checkFileInManifest(t, ts.parseManifest(20, "included"), 20, "/bar")
	checkFileInManifest(t, ts.parseManifest(20, "included-two"), 20, "/baz")

	// Version 30. Add "included-nested" to be included by "included".
	ts.Bundles = append(ts.Bundles, "included-nested")
	ts.copyChroots(20, 30)
	ts.write("image/30/included-nested/foobarbaz", "foobarbaz")
	ts.write("image/30/noship/test-bundle-includes", "included\nincluded-two")
	ts.write("image/30/noship/included-includes", "included-nested")
	ts.createManifests(30)
	ts.createFullfiles(30)

	ts.checkExists("www/30/Manifest.os-core")
	ts.checkNotExists("www/30/Manifest.test-bundle")
	ts.checkNotExists("www/30/Manifest.included-two")

	// Note: original test in swupd-server expected zero packs for all bundles
	// including ones missing manifests, the new code fails in that case, since these
	// packs are not used or generated in practice.
	ts.createPack("os-core", 0, 30, ts.path("image"))
	ts.createPack("included", 0, 30, ts.path("image"))
	ts.createPack("included-nested", 0, 30, ts.path("image"))

	mustValidateZeroPack(t, ts.path("www/30/Manifest.os-core"), ts.path("www/30/pack-os-core-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/30/Manifest.included"), ts.path("www/30/pack-included-from-0.tar"))
	mustValidateZeroPack(t, ts.path("www/30/Manifest.included-nested"), ts.path("www/30/pack-included-nested-from-0.tar"))

	checkIncludes(t, ts.parseManifest(30, "included"), "os-core", "included-nested")
	checkFileInManifest(t, ts.parseManifest(30, "included-nested"), 30, "/foobarbaz")
}

// Imported from swupd-server/test/functional/full-run.
func TestFullRun(t *testing.T) {
	ts := newTestSwupd(t, "full-run-")
	defer ts.cleanup()

	ts.Bundles = []string{"os-core", "test-bundle"}

	ts.write("image/10/test-bundle/foo", "foo")
	ts.createManifests(10)
	ts.createFullfiles(10)

	infoOsCore := ts.createPack("os-core", 0, 10, "")
	mustValidateZeroPack(t, ts.path("www/10/Manifest.os-core"), ts.path("www/10/pack-os-core-from-0.tar"))
	mustHaveDeltaCount(t, infoOsCore, 0)
	// Empty file (bundle file), empty dir, os-release.
	mustHaveFullfileCount(t, infoOsCore, 3)

	infoTestBundle := ts.createPack("test-bundle", 0, 10, "")
	mustValidateZeroPack(t, ts.path("www/10/Manifest.test-bundle"), ts.path("www/10/pack-test-bundle-from-0.tar"))
	mustHaveDeltaCount(t, infoTestBundle, 0)
	// Empty file (bundle file), "foo".
	mustHaveFullfileCount(t, infoTestBundle, 2)

	testBundle := ts.parseManifest(10, "test-bundle")
	checkIncludes(t, testBundle, "os-core")
	checkFileInManifest(t, testBundle, 10, "/usr/share/clear/bundles/test-bundle")

	osCore := ts.parseManifest(10, "os-core")
	checkIncludes(t, osCore)
	checkFileInManifest(t, osCore, 10, "/usr")
	checkFileInManifest(t, osCore, 10, "/usr/lib")
	checkFileInManifest(t, osCore, 10, "/usr/share")
	checkFileInManifest(t, osCore, 10, "/usr/share/clear")
	checkFileInManifest(t, osCore, 10, "/usr/share/clear/bundles")
	checkFileInManifest(t, osCore, 10, "/usr")
}
import json
import os

cases = [
    # Good
    {"id": "G01", "cat": "good", "file": "good_01_distilled_spirits_clean_front.png", "bev": "distilled_spirits", "axis": "clean baseline", "brand": "OLD TOM DISTILLERY", "class_type": "Kentucky Straight Bourbon Whiskey", "abv": "45% Alc./Vol. (90 Proof)", "net": "750 mL", "notes": "Clean baseline distilled spirits label with all required prototype fields visible."},
    {"id": "G02", "cat": "good", "file": "good_02_bourbon_proof_and_abv.png", "bev": "distilled_spirits", "axis": "proof and ABV extraction", "brand": "BARREL HOUSE NO. 7", "class_type": "Straight Bourbon Whiskey", "abv": "50% Alc./Vol. (100 Proof)", "net": "750 mL", "notes": "Clean label, ABV and proof on same line, high contrast"},
    {"id": "G03", "cat": "good", "file": "good_03_vodka_metric_net_contents.png", "bev": "distilled_spirits", "axis": "metric net contents", "brand": "SILVER BRIDGE VODKA", "class_type": "Vodka", "abv": "40% Alc./Vol. (80 Proof)", "net": "1 L", "notes": "Minimalist white label, crisp sans-serif text"},
    {"id": "G04", "cat": "good", "file": "good_04_wine_abv_decimal.png", "bev": "wine", "axis": "decimal ABV", "brand": "PINE RIDGE CELLARS", "class_type": "Red Wine", "abv": "13.5% Alc./Vol.", "net": "750 mL", "notes": "Elegant but readable wine label"},
    {"id": "G05", "cat": "good", "file": "good_05_beer_abv_low_value.png", "bev": "malt_beverage", "axis": "lower ABV beer label", "brand": "RIVER MALT BREWING", "class_type": "Lager Beer", "abv": "5.2% Alc./Vol.", "net": "12 fl oz", "notes": "Beer can label style"},
    {"id": "G06", "cat": "good", "file": "good_06_import_country_origin.png", "bev": "distilled_spirits", "axis": "country of origin present", "brand": "CASK & COMPASS", "class_type": "Blended Scotch Whisky", "abv": "43% Alc./Vol. (86 Proof)", "net": "750 mL", "notes": "Classic spirits import label"},
    {"id": "G07", "cat": "good", "file": "good_07_brand_case_variation.png", "bev": "distilled_spirits", "axis": "fuzzy brand casing", "brand": "Stone's Throw Spirits", "class_type": "Rye Whiskey", "abv": "46% Alc./Vol. (92 Proof)", "net": "750 mL", "notes": "Brand appears as uppercase on label"},
    {"id": "G08", "cat": "good", "file": "good_08_brand_punctuation_variation.png", "bev": "distilled_spirits", "axis": "punctuation normalization", "brand": "Cask and Compass", "class_type": "American Single Malt Whiskey", "abv": "47% Alc./Vol. (94 Proof)", "net": "750 mL", "notes": "Label shows & instead of and"},
    {"id": "G09", "cat": "good", "file": "good_09_warning_on_back_label.png", "bev": "wine", "axis": "warning lower/back label placement", "brand": "NORTHSTAR BOTTLING CO.", "class_type": "White Wine", "abv": "12.8% Alc./Vol.", "net": "750 mL", "notes": "Back-label style with dense text"},
    {"id": "G10", "cat": "good", "file": "good_10_dense_but_readable_label.png", "bev": "distilled_spirits", "axis": "dense readable label", "brand": "BLACK MAPLE RESERVE", "class_type": "Dark Rum", "abv": "42% Alc./Vol. (84 Proof)", "net": "750 mL", "notes": "Dense label with many text blocks"},

    # Mediocre
    {"id": "M01", "cat": "mediocre", "file": "mediocre_01_slight_blur.jpg", "bev": "distilled_spirits", "axis": "mild blur", "brand": "OLD TOM DISTILLERY", "class_type": "Kentucky Straight Bourbon Whiskey", "abv": "45% Alc./Vol. (90 Proof)", "net": "750 mL", "notes": "Slightly blurred", "status": "Needs Review"},
    {"id": "M02", "cat": "mediocre", "file": "mediocre_02_strong_blur.jpg", "bev": "distilled_spirits", "axis": "stronger blur", "brand": "BARREL HOUSE NO. 7", "class_type": "Straight Bourbon Whiskey", "abv": "50% Alc./Vol. (100 Proof)", "net": "750 mL", "notes": "Noticeably blurred", "status": "Needs Review"},
    {"id": "M03", "cat": "mediocre", "file": "mediocre_03_slight_rotation.jpg", "bev": "distilled_spirits", "axis": "skew", "brand": "SILVER BRIDGE VODKA", "class_type": "Vodka", "abv": "40% Alc./Vol. (80 Proof)", "net": "1 L", "notes": "10 degree rotation", "status": "Needs Review"},
    {"id": "M04", "cat": "mediocre", "file": "mediocre_04_perspective_angle.jpg", "bev": "wine", "axis": "perspective distortion", "brand": "PINE RIDGE CELLARS", "class_type": "Red Wine", "abv": "13.5% Alc./Vol.", "net": "750 mL", "notes": "Perspective trapezoid distortion", "status": "Needs Review"},
    {"id": "M05", "cat": "mediocre", "file": "mediocre_05_low_contrast.png", "bev": "malt_beverage", "axis": "low contrast", "brand": "RIVER MALT BREWING", "class_type": "Lager Beer", "abv": "5.2% Alc./Vol.", "net": "12 fl oz", "notes": "Beige background with light gray text", "status": "Needs Review"},
    {"id": "M06", "cat": "mediocre", "file": "mediocre_06_glare_over_warning.jpg", "bev": "distilled_spirits", "axis": "glare over warning", "brand": "CASK & COMPASS", "class_type": "Blended Scotch Whisky", "abv": "43% Alc./Vol. (86 Proof)", "net": "750 mL", "notes": "Light glare crosses warning", "status": "Needs Review"},
    {"id": "M07", "cat": "mediocre", "file": "mediocre_07_shadow_over_abv.jpg", "bev": "distilled_spirits", "axis": "shadow over alcohol content", "brand": "Stone's Throw Spirits", "class_type": "Rye Whiskey", "abv": "46% Alc./Vol. (92 Proof)", "net": "750 mL", "notes": "Shadow darkens ABV area", "status": "Needs Review"},
    {"id": "M08", "cat": "mediocre", "file": "mediocre_08_small_warning_text.png", "bev": "distilled_spirits", "axis": "tiny warning", "brand": "Cask and Compass", "class_type": "American Single Malt Whiskey", "abv": "47% Alc./Vol. (94 Proof)", "net": "750 mL", "notes": "Warning text is very small", "status": "Needs Review"},
    {"id": "M09", "cat": "mediocre", "file": "mediocre_09_curved_bottle_photo.jpg", "bev": "wine", "axis": "curved bottle photo", "brand": "NORTHSTAR BOTTLING CO.", "class_type": "White Wine", "abv": "12.8% Alc./Vol.", "net": "750 mL", "notes": "Label wrapped on bottle", "status": "Needs Review"},
    {"id": "M10", "cat": "mediocre", "file": "mediocre_10_partial_crop_edges.jpg", "bev": "distilled_spirits", "axis": "partial crop", "brand": "BLACK MAPLE RESERVE", "class_type": "Dark Rum", "abv": "42% Alc./Vol. (84 Proof)", "net": "750 mL", "notes": "Cropped slightly at edges", "status": "Needs Review"},

    # Bad
    {"id": "B01", "cat": "bad", "file": "bad_01_missing_warning.png", "bev": "distilled_spirits", "axis": "missing government warning", "brand": "OLD TOM DISTILLERY", "class_type": "Kentucky Straight Bourbon Whiskey", "abv": "45% Alc./Vol. (90 Proof)", "net": "750 mL", "notes": "Omit government warning", "status": "Likely Fail", "warn": False},
    {"id": "B02", "cat": "bad", "file": "bad_02_wrong_warning_heading_case.png", "bev": "distilled_spirits", "axis": "government warning heading case", "brand": "BARREL HOUSE NO. 7", "class_type": "Straight Bourbon Whiskey", "abv": "50% Alc./Vol. (100 Proof)", "net": "750 mL", "notes": "Change GOVERNMENT WARNING: to Government Warning:", "status": "Likely Fail"},
    {"id": "B03", "cat": "bad", "file": "bad_03_warning_wrong_wording.png", "bev": "distilled_spirits", "axis": "altered warning wording", "brand": "SILVER BRIDGE VODKA", "class_type": "Vodka", "abv": "40% Alc./Vol. (80 Proof)", "net": "1 L", "notes": "Warning wording altered", "status": "Likely Fail"},
    {"id": "B04", "cat": "bad", "file": "bad_04_missing_abv.png", "bev": "wine", "axis": "missing alcohol content", "brand": "PINE RIDGE CELLARS", "class_type": "Red Wine", "abv": "13.5% Alc./Vol.", "net": "750 mL", "notes": "Omit alcohol content", "status": "Likely Fail", "known": {"alcohol_content": None}},
    {"id": "B05", "cat": "bad", "file": "bad_05_abv_mismatch.png", "bev": "distilled_spirits", "axis": "ABV mismatch", "brand": "OLD TOM DISTILLERY", "class_type": "Kentucky Straight Bourbon Whiskey", "abv": "45% Alc./Vol. (90 Proof)", "net": "750 mL", "notes": "Label alcohol content differs", "status": "Likely Fail", "known": {"alcohol_content": "40% Alc./Vol. (80 Proof)"}},
    {"id": "B06", "cat": "bad", "file": "bad_06_missing_net_contents.png", "bev": "distilled_spirits", "axis": "missing net contents", "brand": "CASK & COMPASS", "class_type": "Blended Scotch Whisky", "abv": "43% Alc./Vol. (86 Proof)", "net": "750 mL", "notes": "Omit net contents", "status": "Likely Fail", "known": {"net_contents": None}},
    {"id": "B07", "cat": "bad", "file": "bad_07_brand_mismatch.png", "bev": "distilled_spirits", "axis": "brand mismatch", "brand": "OLD TOM DISTILLERY", "class_type": "Kentucky Straight Bourbon Whiskey", "abv": "45% Alc./Vol. (90 Proof)", "net": "750 mL", "notes": "Brand mismatch", "status": "Likely Fail", "known": {"brand_name": "COPPER FOX HOLLOW"}},
    {"id": "B08", "cat": "bad", "file": "bad_08_class_type_mismatch.png", "bev": "distilled_spirits", "axis": "class/type mismatch", "brand": "BARREL HOUSE NO. 7", "class_type": "Straight Bourbon Whiskey", "abv": "50% Alc./Vol. (100 Proof)", "net": "750 mL", "notes": "Class type mismatch", "status": "Likely Fail", "known": {"class_type": "Vodka"}},
    {"id": "B09", "cat": "bad", "file": "bad_09_unreadably_blurry.jpg", "bev": "wine", "axis": "unreadable image", "brand": "NORTHSTAR BOTTLING CO.", "class_type": "White Wine", "abv": "12.8% Alc./Vol.", "net": "750 mL", "notes": "Extremely blurry", "status": "Needs Review"},
    {"id": "B10", "cat": "bad", "file": "bad_10_extreme_crop_missing_fields.jpg", "bev": "distilled_spirits", "axis": "crop removes required fields", "brand": "BLACK MAPLE RESERVE", "class_type": "Dark Rum", "abv": "42% Alc./Vol. (84 Proof)", "net": "750 mL", "notes": "Crop out bottom half", "status": "Likely Fail", "known": {"net_contents": None, "government_warning_present": False}},
]

manifest = {"cases": []}

for c in cases:
    # manifest entry
    manifest["cases"].append({
        "case_id": c["id"],
        "category": c["cat"],
        "filename": c["file"],
        "expected_overall_status": c.get("status", "Pass")
    })
    
    # expected json
    exp = {
        "case_id": c["id"],
        "filename": c["file"],
        "category": c["cat"],
        "beverage_type": c["bev"],
        "expected_overall_status": c.get("status", "Pass"),
        "primary_test_axis": c["axis"],
        "expected_fields": {
            "brand_name": c["brand"],
            "class_type": c["class_type"],
            "alcohol_content": c["abv"],
            "net_contents": c["net"],
            "government_warning_present": c.get("warn", True)
        },
        "notes": c["notes"]
    }
    if "known" in c:
        exp["known_label_fields"] = c["known"]

    with open(f'samples/expected/{c["cat"]}/{c["file"].replace(".png", ".json").replace(".jpg", ".json")}', 'w') as f:
        json.dump(exp, f, indent=2)

    # prompt
    with open(f'samples/prompts/{c["cat"]}/{c["file"].replace(".png", ".txt").replace(".jpg", ".txt")}', 'w') as f:
        f.write(f"Generate label for {c['brand']}, {c['class_type']}, {c['abv']}, {c['net']}. Axis: {c['axis']}")

with open('samples/manifest.json', 'w') as f:
    json.dump(manifest, f, indent=2)

print("Generated 30 case files and manifest.")

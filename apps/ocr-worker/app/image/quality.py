from PIL import Image, ImageFilter, ImageStat

def calculate_quality(image: Image.Image) -> dict:
    gray = image.convert("L")
    
    stat = ImageStat.Stat(gray)
    stddev = stat.stddev[0]
    contrast_score = min(1.0, stddev / 127.5) 
    
    edges = gray.filter(ImageFilter.FIND_EDGES)
    edge_stat = ImageStat.Stat(edges)
    edge_stddev = edge_stat.stddev[0]
    
    blur_score = min(1.0, edge_stddev / 50.0) 
    
    quality_notes = []
    if blur_score < 0.2:
        quality_notes.append("Image appears highly blurry.")
    if contrast_score < 0.2:
        quality_notes.append("Image has very low contrast.")
        
    return {
        "width": image.width,
        "height": image.height,
        "contrast_score": round(contrast_score, 2),
        "blur_score": round(blur_score, 2),
        "quality_notes": quality_notes
    }

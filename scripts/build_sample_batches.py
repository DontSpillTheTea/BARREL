import os
import zipfile
from pathlib import Path

def create_batch_zip(source_dirs, output_zip, limit=None):
    # Support jpg, jpeg, png
    valid_exts = {'.png', '.jpg', '.jpeg'}
    files_to_zip = []
    
    for src in source_dirs:
        if not os.path.exists(src):
            print(f"Warning: Source directory does not exist: {src}")
            continue
        for file in os.listdir(src):
            if Path(file).suffix.lower() in valid_exts:
                files_to_zip.append(os.path.join(src, file))
    
    if not files_to_zip:
        print(f"No valid images found for {output_zip}. Skipping.")
        return
    
    if limit:
        files_to_zip = files_to_zip[:limit]
        
    os.makedirs(os.path.dirname(output_zip), exist_ok=True)
    with zipfile.ZipFile(output_zip, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for f in files_to_zip:
            arcname = os.path.basename(f)
            zipf.write(f, arcname)
    print(f"Created {output_zip} with {len(files_to_zip)} files.")

def main():
    base_dir = os.path.join(os.path.dirname(__file__), '..', 'samples')
    generated_dir = os.path.join(base_dir, 'generated')
    batches_dir = os.path.join(base_dir, 'batches')
    
    good_src = [os.path.join(generated_dir, 'good')]
    mediocre_src = [os.path.join(generated_dir, 'mediocre')]
    bad_src = [os.path.join(generated_dir, 'bad')]
    mixed_src = good_src + mediocre_src + bad_src
    
    create_batch_zip(good_src, os.path.join(batches_dir, 'good_10.zip'), limit=10)
    create_batch_zip(mediocre_src, os.path.join(batches_dir, 'mediocre_10.zip'), limit=10)
    create_batch_zip(bad_src, os.path.join(batches_dir, 'bad_10.zip'), limit=10)
    create_batch_zip(mixed_src, os.path.join(batches_dir, 'mixed_30.zip'), limit=30)
    
if __name__ == '__main__':
    main()

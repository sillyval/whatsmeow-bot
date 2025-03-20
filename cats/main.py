import os
from PIL import Image
from collections import defaultdict
import time
import os
import hashlib
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

DIR = "/home/oliver/whatsmeow-bot/cats/"

def calculate_file_hash(filepath):
    """Calculate SHA256 hash of the file."""
    hash_sha256 = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            hash_sha256.update(chunk)
    return hash_sha256.hexdigest()

def wait_for_file(filepath, delay=1):
    last_size = -1
    while True:
        current_size = os.path.getsize(filepath)
        if current_size == last_size:
            return True
        last_size = current_size
        time.sleep(delay)

def find_duplicate_images(directory):
    hash_dict = defaultdict(list)
    ignore_dirs = {'venv'}  # Add directories to ignore

    for subdir, dirs, files in os.walk(directory):
        dirs[:] = [d for d in dirs if d not in ignore_dirs]

        for file in files:
            if not file.lower().endswith(('.png', '.jpg', '.jpeg', '.gif', '.bmp')):  # Process only images
                continue
            
            try:
                file_path = os.path.join(subdir, file)
                img_hash = calculate_file_hash(file_path)
                hash_dict[img_hash].append(file_path)
            except Exception as e:
                print(f"Error processing {file_path}: {e}")

    duplicates = {hash_val: paths for hash_val, paths in hash_dict.items() if len(paths) > 1}
    return duplicates

def rename_to_latest(directory, file_path):
    """
    Rename the file to 'cat_NUMBER.jpg' where NUMBER is the next available index.
    Uses 'total.txt' to track the current index, or calculates it if 'total.txt' doesn't exist.
    
    Args:
        directory (str): Directory where the file is located.
        file_path (str): Path to the new file to rename.
    """
    total_file_path = os.path.join(directory, "total.txt")

    # Determine CurrentIndex
    if os.path.exists(total_file_path):
        # Read CurrentIndex from total.txt
        with open(total_file_path, "r") as total_file:
            try:
                current_index = int(total_file.read().strip())
            except ValueError:
                print("Invalid value in total.txt. Resetting to 0.")
                current_index = 0
    else:
        # Calculate CurrentIndex manually
        existing_files = [f for f in os.listdir(directory)
                          if f.startswith("cat_") and f.endswith(".jpg")]

        indices = []
        for filename in existing_files:
            parts = filename.split("_")
            if len(parts) == 2 and parts[1].replace(".jpg", "").isdigit():
                indices.append(int(parts[1].replace(".jpg", "")))

        current_index = max(indices, default=0)  # Highest existing index or 0 if none

    # Increment CurrentIndex for the new file
    current_index += 1
    new_filename = f"cat_{current_index}.jpg"
    new_file_path = os.path.join(directory, new_filename)

    # Ensure the file doesn't overwrite an existing one (safety check)
    while os.path.exists(new_file_path):
        current_index += 1
        new_filename = f"cat_{current_index}.jpg"
        new_file_path = os.path.join(directory, new_filename)

    # Rename and update total.txt
    try:
        os.rename(file_path, new_file_path)
        print(f"File renamed to: {new_file_path}")

        # Write the updated index back to total.txt
        with open(total_file_path, "w") as total_file:
            total_file.write(str(current_index))
        print(f"Updated total.txt to: {current_index}")

    except Exception as e:
        print(f"Error renaming file: {e}")


class FileAddedHandler(FileSystemEventHandler):
    def on_created(self, event):
        if event.is_directory:
            return

        new_file = event.src_path
        # Ignore non-image files
        if not new_file.lower().endswith(('.png', '.jpg', '.jpeg', '.gif', '.bmp')):
            print(f"Ignored non-image file: {new_file}")
            return

        # Ignore files in the 'venv' directory
        if 'venv' in new_file.split(os.sep):
            print(f"Ignored virtual environment file: {new_file}")
            return

        print(f"File detected: {new_file}")

        wait_for_file(new_file)

        directory = os.path.dirname(new_file)
        new_file_hash = calculate_file_hash(new_file)

        for filename in os.listdir(directory):
            file_path = os.path.join(directory, filename)
            if file_path == new_file or not os.path.isfile(file_path):
                continue

            if not filename.lower().endswith(('.png', '.jpg', '.jpeg', '.gif', '.bmp')):
                continue

            existing_file_hash = calculate_file_hash(file_path)
            if existing_file_hash == new_file_hash:
                print(f"Duplicate detected. Deleting {new_file}")
                os.remove(new_file)
                return

        rename_to_latest(directory, new_file)

def monitor_directory(path):
    event_handler = FileAddedHandler()
    observer = Observer()
    observer.schedule(event_handler, path, recursive=False)
    observer.start()
    print(f"Monitoring directory: {path}")

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()

if __name__ == "__main__":
    monitor_directory(DIR)
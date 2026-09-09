#!/usr/bin/env python3
import json
import os
import sys

from PIL import Image


def pack_images(image_paths, output_image_path, output_json_path, max_width=1024):
    """
    A simple shelf-packing algorithm to generate a sprite atlas from a list of images.
    """
    images = []
    for path in image_paths:
        try:
            img = Image.open(path)
            images.append({'path': path, 'image': img, 'w': img.width, 'h': img.height})
        except Exception as e:
            print(f"Failed to load {path}: {e}")

    # Sort by height descending
    images.sort(key=lambda x: x['h'], reverse=True)

    current_x = 0
    current_y = 0
    max_h_in_row = 0
    packed_data = {}

    # Calculate dimensions
    atlas_width = max_width
    atlas_height = 0

    for item in images:
        if current_x + item['w'] > max_width:
            current_y += max_h_in_row
            current_x = 0
            max_h_in_row = 0

        item['x'] = current_x
        item['y'] = current_y

        current_x += item['w']
        if item['h'] > max_h_in_row:
            max_h_in_row = item['h']

    atlas_height = current_y + max_h_in_row

    # Create the atlas
    atlas = Image.new('RGBA', (atlas_width, atlas_height), (0, 0, 0, 0))
    for item in images:
        atlas.paste(item['image'], (item['x'], item['y']))
        packed_data[os.path.basename(item['path'])] = {
            'x': item['x'], 'y': item['y'], 'w': item['w'], 'h': item['h']
        }

    atlas.save(output_image_path)
    with open(output_json_path, 'w') as f:
        json.dump(packed_data, f, indent=4)

    print(f"Atlas generated at {output_image_path} ({atlas_width}x{atlas_height})")

if __name__ == '__main__':
    if len(sys.argv) < 4:
        print("Usage: generate_sprite_atlas.py <output.png> <output.json> <img1.png> [img2.png ...]")
        sys.exit(1)
    pack_images(sys.argv[3:], sys.argv[1], sys.argv[2])

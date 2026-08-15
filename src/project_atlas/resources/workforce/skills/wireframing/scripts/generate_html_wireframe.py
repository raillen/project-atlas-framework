import json
import sys

def generate_wireframe(json_path, output_path):
    with open(json_path, 'r') as f:
        data = json.load(f)
        
    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Wireframe: {data.get('title', 'Page')}</title>
<style>
    body {{ font-family: sans-serif; margin: 0; padding: 20px; background: #f0f0f0; }}
    .wireframe-container {{ max-width: 800px; margin: 0 auto; background: white; border: 2px solid #333; }}
    header, footer {{ background: #ccc; padding: 20px; text-align: center; border-bottom: 2px solid #333; }}
    footer {{ border-bottom: none; border-top: 2px solid #333; }}
    .content {{ padding: 20px; min-height: 400px; }}
    .box {{ border: 2px dashed #999; padding: 20px; margin-bottom: 20px; text-align: center; color: #666; }}
</style>
</head>
<body>
<div class="wireframe-container">
"""
    
    if 'header' in data:
        html += f"<header><h1>{data['header']}</h1></header>\n"
        
    html += "<div class=\"content\">\n"
    for section in data.get('sections', []):
        html += f"  <div class=\"box\">{section['name']}</div>\n"
    html += "</div>\n"
    
    if 'footer' in data:
        html += f"<footer>{data['footer']}</footer>\n"
        
    html += """
</div>
</body>
</html>
"""

    with open(output_path, 'w') as f:
        f.write(html)
    print(f"Wireframe generated at {output_path}")

if __name__ == "__main__":
    # Example usage data
    sample_data = {
        "title": "Dashboard Wireframe",
        "header": "App Logo & Navigation",
        "sections": [
            {"name": "Hero Section (Image Placeholder)"},
            {"name": "Key Metrics (3 Columns)"},
            {"name": "Recent Activity Table"}
        ],
        "footer": "Footer Links & Copyright"
    }
    with open("temp_spec.json", "w") as f:
        json.dump(sample_data, f)
        
    generate_wireframe("temp_spec.json", "wireframe_output.html")

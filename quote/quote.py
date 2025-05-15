import os
import sys
from PIL import Image, ImageDraw, ImageFont

def wrap_text(draw, text, font, max_width):
    words = text.split()
    lines = []
    current_line = []

    for word in words:
        bbox_width = draw.textbbox((0, 0), word, font=font)[2]
        if bbox_width > max_width:
            split_word = []
            while word:
                part = word[:1]
                while draw.textbbox((0, 0), part, font=font)[2] < max_width and len(word) > len(part):
                    part = word[:len(part) + 1]
                split_word.append(part)
                word = word[len(part):]
            lines.extend(split_word)
        else:
            current_line.append(word)
            candidate_line = ' '.join(current_line)
            bbox_width = draw.textbbox((0, 0), candidate_line, font=font)[2]

            if bbox_width <= max_width:
                continue

            current_line.pop()
            lines.append(' '.join(current_line))
            current_line = [word]

    if current_line:
        lines.append(' '.join(current_line))
    return lines

def create_quote_image(name, body, image_filepath):

    body = f"\"{body}\""

    img = Image.new('RGB', (1300, 600), color=(0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    if os.path.exists(image_filepath):
        profile_picture = Image.open(image_filepath)
        profile_picture = profile_picture.resize((600, 600), Image.LANCZOS)
        img.paste(profile_picture, (0, 0))
    
    gradient = Image.new('RGBA', (700, 600))
    for x in range(700):
        opacity = int(255 * (x / 700))
        for y in range(600):
            gradient.putpixel((x, y), (0, 0, 0, opacity))
    img.paste(gradient, (-100, 0), gradient)

    font_path = os.path.join(os.path.dirname(__file__), "NotoSans-VariableFont_wdth,wght.ttf")
    name_font_path = os.path.join(os.path.dirname(__file__), "NotoSans-Italic-VariableFont_wdth,wght.ttf")

    body_font = ImageFont.truetype(font_path, 30)
    name_font = ImageFont.truetype(name_font_path, 24)

    wrapped_text = wrap_text(draw, body, body_font, max_width=550)
    y = (600 - (len(wrapped_text) * 40)) // 2

    for line in wrapped_text:
        text_width = draw.textbbox((0, 0), line, font=body_font)[2]
        draw.text(((950 - text_width / 2), y), line, font=body_font, fill="#FFFFFF")
        y += 40

    username_text = f"~{name}"
    text_width = draw.textbbox((0, 0), username_text, font=name_font)[2]
    draw.text(((950 - text_width / 2), y + 10), username_text, font=name_font, fill="#BBBBBB")
    
    output_path = os.path.join(os.path.dirname(__file__), f"{".".join(image_filepath.split("/")[-1].split(".")[0:-1])}_output.png")
    img.save(output_path)
    print(f"Saved quote image to {output_path}")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: python generate_quote.py <name> <body> <image_filepath>")
        sys.exit(1)

    name = sys.argv[1]
    body = sys.argv[2]
    image_filepath = sys.argv[3]
    create_quote_image(name, body, image_filepath)

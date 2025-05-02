import base64
import sys
import os

# filename = "my_image.png" # Or get from argv
if len(sys.argv) < 2:
    print("Usage: python im.py <image_path>")
    sys.exit(1)

filename = sys.argv[1]
with open(filename, "rb") as f:
    data = f.read()
b64data = base64.b64encode(data).decode('ascii')
# Get size for optional width/height hints
# size_bytes = os.path.getsize(filename)
# print(f"\x1b]1337;File=inline=1;size={size_bytes}:{b64data}\a")

# Basic inline image sequence
# Options like width, height, preserveAspectRatio can be added, e.g.,
# print(f"\x1b]1337;File=inline=1;width=100%;preserveAspectRatio=1:{b64data}\a")
print(f"\x1b]1337;File=inline=1:{b64data}\a")
sys.stdout.flush() 
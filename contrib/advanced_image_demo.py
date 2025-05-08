#!/usr/bin/env python3
"""
GT Terminal Advanced Image Demo
-------------------------------
This script demonstrates the advanced image rendering capabilities
of GT Terminal including z-index, alignment, and positioning.
"""

import argparse
import base64
import os
import sys
import time
from pathlib import Path
from typing import Dict, Any, Optional, List, Tuple

def display_image(path: str, options: Dict[str, Any]) -> None:
    """Display an image in the terminal using the GT/iTerm2 protocol"""
    if not os.path.isfile(path):
        print(f"Error: Image file not found: {path}", file=sys.stderr)
        return

    with open(path, "rb") as f:
        image_data = base64.b64encode(f.read()).decode("ascii")

    # Build options string (always include inline=1)
    opts = ["inline=1"]
    for key, value in options.items():
        opts.append(f"{key}={value}")

    # Output the escape sequence
    sys.stdout.write(f"\033]1337;File={';'.join(opts)}:{image_data}\a")
    sys.stdout.flush()

def clear_screen() -> None:
    """Clear the terminal screen"""
    sys.stdout.write("\033[2J\033[H")
    sys.stdout.flush()

def move_cursor(row: int, col: int) -> None:
    """Move cursor to specified position"""
    sys.stdout.write(f"\033[{row};{col}H")
    sys.stdout.flush()

def demo_basic(image_dir: Path) -> None:
    """Basic image display demo"""
    clear_screen()
    print("GT Terminal Advanced Image Demo - Basic Display")
    print("==============================================")
    print()
    
    # Find sample images
    sample_img = next(image_dir.glob("*.jpg"), None) or next(image_dir.glob("*.png"), None)
    if not sample_img:
        print("No sample images found in the directory.")
        return
    
    # Display with different sizes
    print("1. Default size (auto):")
    display_image(str(sample_img), {})
    print("\n" * 8)  # Space for the image
    
    print("2. Fixed width (400px):")
    display_image(str(sample_img), {"width": "400px"})
    print("\n" * 8)  # Space for the image
    
    print("3. Fixed height (200px):")
    display_image(str(sample_img), {"height": "200px"})
    print("\n" * 8)  # Space for the image

    print("4. Percentage width (50%):")
    display_image(str(sample_img), {"width": "50%"})
    print("\n" * 8)  # Space for the image

    print("Press Enter to continue...")
    input()

def demo_alignment(image_dir: Path) -> None:
    """Alignment demo"""
    clear_screen()
    print("GT Terminal Advanced Image Demo - Alignment")
    print("=========================================")
    print()
    
    # Find sample images
    sample_img = next(image_dir.glob("*.jpg"), None) or next(image_dir.glob("*.png"), None)
    if not sample_img:
        print("No sample images found in the directory.")
        return
    
    # Display with different alignments
    print("1. Left alignment (default):")
    display_image(str(sample_img), {"width": "200px", "align": "left"})
    print("\n" * 5)  # Space for the image
    
    print("2. Center alignment:")
    display_image(str(sample_img), {"width": "200px", "align": "center"})
    print("\n" * 5)  # Space for the image
    
    print("3. Right alignment:")
    display_image(str(sample_img), {"width": "200px", "align": "right"})
    print("\n" * 5)  # Space for the image
    
    print("Press Enter to continue...")
    input()

def demo_z_index(image_dir: Path) -> None:
    """Z-index demo with overlapping images"""
    clear_screen()
    print("GT Terminal Advanced Image Demo - Z-index Layering")
    print("================================================")
    print()
    
    # Find sample images
    images = list(image_dir.glob("*.jpg")) + list(image_dir.glob("*.png"))
    if len(images) < 2:
        print("Need at least 2 images for the z-index demo.")
        return
    
    img1, img2 = images[0], images[1]
    
    print("Displaying overlapping images with different z-index values:")
    print("(Notice how they stack on top of each other)")
    print()

    move_cursor(6, 5)
    display_image(str(img1), {
        "width": "40%", 
        "z-index": "1", 
        "name": "back_image",
    })
    
    # Pause briefly to let the first image render
    time.sleep(0.5)
    
    move_cursor(8, 15)  # Position for second image overlapping first
    display_image(str(img2), {
        "width": "40%", 
        "z-index": "2", 
        "name": "front_image",
    })
    
    move_cursor(15, 1)
    print("Second image (z-index=2) appears on top of the first image (z-index=1)")
    print()
    print("Press Enter to continue...")
    input()

def demo_animation(image_dir: Path) -> None:
    """Simple animation demo using z-index and positioning"""
    clear_screen()
    print("GT Terminal Advanced Image Demo - Animation")
    print("=========================================")
    print("Press Ctrl+C to stop the animation")
    print()
    
    # Find sample image
    sample_img = next(image_dir.glob("*.jpg"), None) or next(image_dir.glob("*.png"), None)
    if not sample_img:
        print("No sample images found in the directory.")
        return
    
    try:
        # Create a persistent background
        move_cursor(5, 1)
        display_image(str(sample_img), {
            "width": "80%", 
            "height": "80%",
            "align": "center",
            "z-index": "1", 
            "name": "background",
            "persistent": "1"
        })
        
        # Display an animated overlay
        for i in range(100):
            # Bouncing animation pattern
            row = 7 + abs((i % 20) - 10)
            col = 10 + i % 40
            
            move_cursor(row, col)
            display_image(str(sample_img), {
                "width": "100px",
                "height": "100px",
                "z-index": "2",  
                "name": f"animation_frame",
            })
            
            # Wait a bit before next frame
            time.sleep(0.1)
            
    except KeyboardInterrupt:
        print("\nAnimation stopped.")
    
    move_cursor(25, 1)
    print("Press Enter to continue...")
    input()

def main() -> None:
    parser = argparse.ArgumentParser(description="GT Terminal Advanced Image Demo")
    parser.add_argument("--image-dir", type=str, default=".", 
                        help="Directory containing sample images (JPG/PNG)")
    
    args = parser.parse_args()
    image_dir = Path(args.image_dir)
    
    if not image_dir.is_dir():
        print(f"Error: Image directory not found: {image_dir}", file=sys.stderr)
        sys.exit(1)
        
    # Check for images
    image_count = len(list(image_dir.glob("*.jpg")) + list(image_dir.glob("*.png")))
    if image_count == 0:
        print(f"Error: No JPG or PNG images found in {image_dir}", file=sys.stderr)
        sys.exit(1)
    
    print(f"Found {image_count} images in {image_dir}")
    
    try:
        demo_basic(image_dir)
        demo_alignment(image_dir)
        demo_z_index(image_dir)
        demo_animation(image_dir)
        
        clear_screen()
        print("GT Terminal Advanced Image Demo - Complete")
        print("========================================")
        print("Thank you for trying the advanced image features!")
        
    except Exception as e:
        print(f"Demo error: {str(e)}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
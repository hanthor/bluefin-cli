# Design the raptor as an 8-row pixel grid (2 braille rows), modeled on the
# Chrome offline T-rex: boxy head with flat snout + eye notch, thick body,
# tail rising to the left, tiny arm, two legs. Prints the pixel art and the
# braille conversion for visual comparison.

def to_braille(grid):
    rows = len(grid)
    cols = max(len(r) for r in grid)
    grid = [r.ljust(cols) for r in grid]
    bits = {(0,0):0x01,(0,1):0x02,(0,2):0x04,(0,3):0x40,
            (1,0):0x08,(1,1):0x10,(1,2):0x20,(1,3):0x80}
    out = []
    for by in range(0, rows, 4):
        line = ""
        for bx in range(0, cols, 2):
            v = 0
            for dx in range(2):
                for dy in range(4):
                    y, x = by+dy, bx+dx
                    if y < rows and x < cols and grid[y][x] == '#':
                        v |= bits[(dx,dy)]
            line += chr(0x2800+v)
        out.append(line)
    return out

# 16 cols x 16 rows, hand-refined from a box-filter downsample of the real
# Chrome offline T-rex sprite (tmp/dino_run*.png from the t-rex-runner
# assets): solid head with eye notch, open jaw, tail sweeping down-left,
# arm nub, alternating legs. Rows 14-15 are the leg rows.
BODY = [
    "........########",
    "........##.#####",  # eye (blank pixel)
    "........########",
    "........########",
    "........#####...",
    "........####....",  # open jaw
    "........######..",
    ".#....######....",  # tail tip
    ".##..#######.#..",  # arm nub
    ".###########.#..",
    ".###########....",
    ".##########.....",
    "..#########.....",
    "...########.....",
]
LEGS = {
    "run A": ["....##...##.....", "....#....#......"],
    "run B": ["....##..##......", ".....##...#....."],
    "pause": ["....##...##.....", "....##...##....."],
}

for name, legs in LEGS.items():
    grid = BODY + legs
    print(f"--- {name} ---")
    for r in grid:
        print("  " + r.replace('.', '\u00b7').replace('#', '\u2588'))
    for line in to_braille(grid):
        print("  |" + line + "|")

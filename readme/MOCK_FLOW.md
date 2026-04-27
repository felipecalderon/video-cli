# Mock buffers and flow simulation (fake data)

Scenario: `termW=2`, `termH=2` => `WorkRGB.H=4` (vertical subcells).

## 1) FrameRGB (fake input)

`FrameRGB` 4x4 (interleaved RGB, sRGB):

Row 0: (10,20,30) (40,50,60) (70,80,90) (100,110,120)
Row 1: (12,22,32) (42,52,62) (72,82,92) (102,112,122)
Row 2: (14,24,34) (44,54,64) (74,84,94) (104,114,124)
Row 3: (16,26,36) (46,56,66) (76,86,96) (106,116,126)

## 2) ResizeToTerm -> WorkRGB 2x4

Downscale with a simple sampling (mock):

Row 0: (20,30,40) (90,100,110)
Row 1: (22,32,42) (92,102,112)
Row 2: (24,34,44) (94,104,114)
Row 3: (26,36,46) (96,106,116)

## 3) Quantize (6bpc) + Dither (Bayer 4x4)

Mock result (already dithered):

Row 0: (20,32,40) (88,104,112)
Row 1: (24,32,40) (96,104,112)
Row 2: (24,32,48) (96,112,120)
Row 3: (24,40,48) (96,112,120)

## 4) MapCells -> CellGrid 2x2

Each cell uses 2 vertical subcells:

Cell (0,0):
- Top = Row0 Col0 -> (20,32,40)
- Bottom = Row1 Col0 -> (24,32,40)
- Ch = '▀'

Cell (1,0):
- Top = Row0 Col1 -> (88,104,112)
- Bottom = Row1 Col1 -> (96,104,112)
- Ch = '▀'

Cell (0,1):
- Top = Row2 Col0 -> (24,32,48)
- Bottom = Row3 Col0 -> (24,40,48)
- Ch = '▀'

Cell (1,1):
- Top = Row2 Col1 -> (96,112,120)
- Bottom = Row3 Col1 -> (96,112,120)
- Ch = '▀'

## 5) Diff (prev vs actual)

Prev `CellGrid` (mock, only cell (1,0) changed):

Cell (1,0) before:
- Top = (80,96,104)
- Bottom = (88,96,104)
- Ch = '▀'

Diff result: emit only (1,0) and its colors.

## 6) Output (conceptual example)

- Move cursor to (x=1, y=0)
- Set FG to Top color, BG to Bottom color
- Print '▀'


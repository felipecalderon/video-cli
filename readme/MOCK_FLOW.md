# Mock de buffers y simulacion de flujo (datos falsos)

Escenario: `termW=2`, `termH=2` => `WorkRGB.H=4` (subceldas verticales).

## 1) FrameRGB (entrada fake)

`FrameRGB` 4x4 (RGB interleaved, sRGB):

Fila 0: (10,20,30) (40,50,60) (70,80,90) (100,110,120)  
Fila 1: (12,22,32) (42,52,62) (72,82,92) (102,112,122)  
Fila 2: (14,24,34) (44,54,64) (74,84,94) (104,114,124)  
Fila 3: (16,26,36) (46,56,66) (76,86,96) (106,116,126)

## 2) ResizeToTerm -> WorkRGB 2x4

Downscale por muestreo simple (mock):

Fila 0: (20,30,40) (90,100,110)  
Fila 1: (22,32,42) (92,102,112)  
Fila 2: (24,34,44) (94,104,114)  
Fila 3: (26,36,46) (96,106,116)

## 3) Quantize (6bpc) + Dither (Bayer 4x4)

Resultado mock (ya dithered):

Fila 0: (20,32,40) (88,104,112)  
Fila 1: (24,32,40) (96,104,112)  
Fila 2: (24,32,48) (96,112,120)  
Fila 3: (24,40,48) (96,112,120)

## 4) MapCells -> CellGrid 2x2

Cada celda usa 2 subceldas verticales:

Celda (0,0):
- Top = Fila0 Col0 -> (20,32,40)
- Bottom = Fila1 Col0 -> (24,32,40)
- Ch = '▀'

Celda (1,0):
- Top = Fila0 Col1 -> (88,104,112)
- Bottom = Fila1 Col1 -> (96,104,112)
- Ch = '▀'

Celda (0,1):
- Top = Fila2 Col0 -> (24,32,48)
- Bottom = Fila3 Col0 -> (24,40,48)
- Ch = '▀'

Celda (1,1):
- Top = Fila2 Col1 -> (96,112,120)
- Bottom = Fila3 Col1 -> (96,112,120)
- Ch = '▀'

## 5) Diff (prev vs actual)

Prev `CellGrid` (mock, solo cambia la celda (1,0)):

Celda (1,0) antes:
- Top = (80,96,104)
- Bottom = (88,96,104)
- Ch = '▀'

Resultado diff: se emite solo (1,0) y sus colores.

## 6) Output (ejemplo conceptual)

- Mover cursor a (x=1, y=0)
- Set FG a Top, BG a Bottom
- Print '▀'


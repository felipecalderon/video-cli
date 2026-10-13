# AGENTS.md

## 1. Reglas generales

- Realiza únicamente los cambios solicitados explícitamente por el usuario.
- No amplíes el alcance, refactorices código ajeno a la tarea ni agregues funcionalidades no solicitadas.
- Inspecciona únicamente los archivos de código y documentación necesarios para resolver la tarea. No inspecciones tests ni infraestructura de pruebas por iniciativa propia.
- **Nunca ejecutes tests, compilaciones, builds, benchmarks, linters, `go vet` ni otros comandos de validación.** La validación será realizada manualmente por el usuario.
- **Nunca crees commits ni ejecutes operaciones de Git que modifiquen el estado del repositorio.** No hagas `git add`, `git commit`, cambios de rama, resets ni operaciones equivalentes.
- No sobrescribas, elimines ni reviertas cambios existentes del usuario.
- No modifiques dependencias, configuraciones, interfaces públicas ni contratos existentes sin una justificación clara.
- Si los requisitos son ambiguos o la solución implica decisiones arquitectónicas importantes, consulta al usuario antes de proceder.

## 2. Estándares de Go

- Sigue las convenciones idiomáticas de Go y prioriza la simplicidad, legibilidad y mantenibilidad.
- Utiliza nombres claros, funciones cohesivas y estructuras de control explícitas.
- Evita abstracciones innecesarias, generalizaciones prematuras, estado global mutable y sobreingeniería.
- Gestiona los errores explícitamente. No los ignores silenciosamente y utiliza `%w` cuando corresponda para envolver errores.
- Prefiere la biblioteca estándar y las dependencias existentes. No agregues dependencias nuevas sin una necesidad justificada.
- Respeta la cancelación mediante `context.Context` cuando corresponda.
- Gestiona correctamente los recursos: cierra archivos, procesos, pipes y demás recursos de forma determinista.
- En código concurrente, evita condiciones de carrera, bloqueos, fugas de goroutines y sincronización innecesaria.
- Respeta el formato convencional de Go. No ejecutes herramientas de formato o validación automáticamente; el usuario se encargará de verificarlo.
- Optimiza únicamente cuando exista una necesidad real. Evita asignaciones, bloqueos y operaciones innecesarias en rutas críticas de rendimiento.

## 3. Prioridades específicas de VTerminal

VTerminal es un motor de renderizado de video en tiempo real dentro de la terminal, con sincronización de audio, colores ANSI truecolor y técnicas de dithering.

Al modificar el proyecto, prioriza lo siguiente:

1. **Sincronización audiovisual:** conserva la correspondencia temporal entre audio y video. Evita soluciones que introduzcan desfases, saltos o acumulación progresiva de latencia.
2. **Rendimiento en tiempo real:** minimiza bloqueos, asignaciones innecesarias y trabajo redundante durante el procesamiento de cada frame.
3. **Renderizado incremental:** preserva el sistema de comparación de frames y la escritura diferencial para reducir la salida enviada a la terminal.
4. **Calidad visual:** conserva el comportamiento del truecolor, el dithering perceptual, la cuantización y la representación mediante caracteres.
5. **Compatibilidad:** evita cambios incompatibles en la CLI, la configuración, los formatos de entrada y el comportamiento existente.
6. **Gestión de procesos externos:** trata FFmpeg y yt-dlp como dependencias externas susceptibles a fallos. Gestiona errores, cancelaciones, terminación de procesos y liberación de recursos correctamente.

Evita introducir operaciones bloqueantes, buffers sin límites o procesamiento redundante en las rutas críticas de reproducción.

## 4. Arquitectura y cambios

- Respeta la arquitectura existente y consulta `readme/ARCHITECTURE.md` cuando sea necesario.
- Consulta `readme/Configuration.md` si la tarea afecta a opciones de configuración.
- Mantén las responsabilidades de decodificación, procesamiento de frames, renderizado, reproducción de audio y sincronización tan separadas como permita el diseño actual.
- Prefiere la solución más pequeña que resuelva correctamente el problema.
- No rediseñes componentes existentes si un cambio localizado es suficiente.
- No inventes APIs, supongas comportamientos no documentados ni cambies silenciosamente la semántica existente.
- Si una modificación puede afectar a la sincronización, el rendimiento, la compatibilidad o la arquitectura, identifica el riesgo antes de implementarla.
- No agregues documentación, archivos auxiliares ni configuraciones que no sean necesarios para la tarea.

## 5. Validación y entrega

- No ejecutes pruebas ni comandos de validación, aunque consideres que son necesarios para confirmar la solución.
- No afirmes que el proyecto compila, que las pruebas pasan o que un comportamiento funciona si no existe evidencia verificable.
- Al finalizar, resume brevemente:
  - Qué archivos modificaste.
  - Qué cambios realizaste y por qué.
  - Qué riesgos o efectos secundarios relevantes identificaste.
  - Qué pasos debería ejecutar manualmente el usuario para validar el resultado, cuando corresponda.

**La implementación y la validación son responsabilidades separadas: el agente realiza los cambios solicitados y el usuario decide cuándo y cómo verificarlos.**

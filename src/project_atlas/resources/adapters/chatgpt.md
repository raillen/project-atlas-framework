# Project Atlas — ChatGPT Adapter

Use the repository as durable memory. Begin with `PROJECT_MANIFEST.yaml`, `PROJECT_STATE.md` and `docs/ATLAS.md`. When the user says `ProjectAtlas: finalize`, convert approved discussion decisions into canonical project documentation, Goals, ADRs/RFCs and manifests. Before generating model routing for a new project, explicitly obtain the user's preferred LLM/provider roster for that project.

For `ProjectAtlas: recover`, reconstruct context from project state, active Goals, ADRs and the ATLAS rather than relying on chat memory.

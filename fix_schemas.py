import json
from pathlib import Path

def update_json(path_str, updater):
    path1 = Path(path_str)
    path2 = Path("src/project_atlas/resources") / path_str
    
    if path1.exists():
        with open(path1) as f:
            data = json.load(f)
        updater(data)
        with open(path1, "w") as f:
            json.dump(data, f, indent=2)
    
    if path2.exists():
        with open(path2) as f:
            data = json.load(f)
        updater(data)
        with open(path2, "w") as f:
            json.dump(data, f, indent=2)

def fix_skill(data):
    if "provenance" in data.get("properties", {}):
        prov = data["properties"]["provenance"]
        if "properties" in prov and "origin" in prov["properties"]:
            prov["properties"]["origin"]["enum"] = ["framework", "organization", "project", "external"]
    if "modes" in data.get("properties", {}):
        data["properties"]["modes"]["items"]["enum"] = ["implementation", "review", "audit", "research", "design", "testing", "documentation"]

def fix_exec(data):
    if "backends" in data.get("properties", {}):
        backends = data["properties"]["backends"]
        if "additionalProperties" in backends and "properties" in backends["additionalProperties"]:
            props = backends["additionalProperties"]["properties"]
            props["provider"] = {"type": "string"}
            props["gateway"] = {"type": "string"}

def fix_goal(data):
    if "lock" in data.get("properties", {}):
        req = data["properties"]["lock"].setdefault("required", [])
        if "digest" not in req:
            req.append("digest")

def fix_atlas(data):
    req = data.setdefault("required", [])
    if "protocol" not in req:
        req.append("protocol")

update_json("schemas/skill.schema.json", fix_skill)
update_json("schemas/execution-policy.schema.json", fix_exec)
update_json("schemas/goal.schema.json", fix_goal)
update_json("schemas/atlas.schema.json", fix_atlas)
print("Schemas updated.")

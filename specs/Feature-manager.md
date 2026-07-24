# Feature-Manager

in this skill, ai read the spec file/files in md format that user give it.

based on the spec file/files, this skill create `**goals**` and `**tasts**`.

---

## Dictionary

### GOAL

`goal` is the general concept that contain of some tasks that relate to each other. each `goal` should not too big or small.
eahc `goal` should contains 3 tasks at least.

### TASK

`task` is the smallest job that do the specefic section of the spec file.

---

## Algorithm

it should follow this algorithm for generate tasks:

1. check `**graphify**` or `**codebase-memory-mcp**` AI skills. if none of them exists, then stop and ask user to install one of them
2. create md file in the `/tmp` directory. file name should has `FEATURE-Manager-{project_name}-{datetime_in_timestamp}.md` format.
3. read spec file/files.
4. create multiple `**goal**` in the `FEATURE-Manager-{project_name}-{datetime_in_timestamp}.md` file.
   1. each goal should contian description in one or two line.
5. check for each goal:
   1. if it can break to two diffrent goal, then break it
6. check for each goal:
   1. if it's too small or can merge to another goal, them merge them
7. if number of merged goals at least one, them go to step 5
8. based on the spec file/files, generate tasks
9. check for each task:
   1. if it can be break to two tasks, then do it.
10. check all tasks and relate each task to one goal
11. if the goal exist with any task, remove it
12. review tasks and goald and compaired them with given spec file/files. if some section missing, then go to step 4. 

---

## `FEATURE-Manager-{project_name}-{datetime_in_timestamp}.md` Format

this is the format of generated file for feature.

```markdown
# TITLE OF FEATURE
description of what should these goals do and what is the feaure.

## ROADMAP

### GOALS #1
description of the goal.

#### TASKS
- [ ] task NO. 1
- [ ] task NO. 2
```


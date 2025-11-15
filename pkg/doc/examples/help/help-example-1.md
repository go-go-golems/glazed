---
Title: Show the list of all toplevel topics
Slug: help-example-1
Short: |
  ```
  glaze help --list
  ```
Topics:
- help-system
Commands:
- help
Flags:
- list
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can ask the help system to list all toplevel topics (not just the default ones) in 
a concise list.

---

```
❯ glaze help --list                    

   glaze - glaze is a tool to format structured data                                                   
                                                                                                       
  ## General topics                                                                                    
                                                                                                       
  Run  glaze help <topic>  to view a topic's page.                                                     
                                                                                                       
  • help-system - Help System                                                                          
  • templates - Templates                                                                              
                                                                                                       
  ## Examples                                                                                          
                                                                                                       
  Run  glaze help <example>  to view an example in full.                                               
                                                                                                       
  • templates-example-1 - Use a single template for single field output                                
                                                                                                       
  ## Applications                                                                                      
                                                                                                       
  Run  glaze help <application>  to view an application in full.                                       
                                                                                                       
  • exposing-a-simple-sql-table - Exposing a simple SQL table using glaze                              
                                                                                                       
  ## Tutorials                                                                                         
                                                                                                       
  Run  glaze help <tutorial>  to view a tutorial in full.                                              
                                                                                                       
  • a-simple-table-cli - Creating a simple CLI application with glaze                                  
  • a-simple-help-system - Creating a help system for a CLI application with glaze                     
```

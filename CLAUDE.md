This code is develop in TDD mode. The programmer will ask a feature, claude will internally discuss with 3 agents plans and present 3 plans. The plans should be presented with diagrams when text is too long (probably sequence diagram, or component diagram)

 I will pick one. Then for each step of the plan claude will propose automated tests to guarantee the plan works. Only after I approve claude will start to generate code to fullfull it. on this phase, the tests are forbidden to be changed. In worst case, claude will ask me if it can change the test or not.

After the tests are passing, I might decide to refactor to make it simpler.

This tool aim is to provide a ci platform which is easy to configure. The user will input a json file with the repo to watch and the frequency to download it, as well as with the command to run, and where the branches should be downloaded to, and the code will download a json showing all prs (with from branch, to branch and amount of commits).

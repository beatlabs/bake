from diagrams import Cluster, Diagram
from diagrams.programming.language import Go
from diagrams.onprem.container import Docker
from diagrams.custom import Custom
from diagrams.k8s.ecosystem import Helm
import os

scriptPath = os.path.realpath(__file__)
scriptDir= os.path.dirname(scriptPath)

with Diagram(name="Bake Architecture", filename="bake", show=False):
    with Cluster(""):
        bake = Docker("Bake container")
        with Cluster(""):
            mage = Go("Mage")
            diagrams = Custom("diagrams - system architecture in Python", os.path.join(scriptDir, "diagrams.png"))
            hadolint = Custom("hadolint - Dockerfile linter", os.path.join(scriptDir, "haskell.png"))
            protobuf = Custom("protoc - protobuf compiler", os.path.join(scriptDir, "protobuf.png"))
            swag = Go("swag - Go annotations to Swagger Documentation")
            mark = Go("Mark - Confluence doc sync")
            helm = Helm("Helm - Helm chart linting")
            golangci = Go("golangci-lint - Fast Go linters runner")
            skim = Go("Skim - Beat internal schema breaking changes detector")

        #

#     with Cluster("Sonar"):
#         sonar = Go()
#         rest >> sonar
#         aws >> sonar
#         sns >> sonar
#         vehicleServices >> sonar
#         driverPositions >> sonar
#         partnerShifts >> sonar
#         fleet >> sonar
#         dracarys >> sonar
#         sonar >> fraudPrevention
#         sonar >> audit
#         sonar >> eta

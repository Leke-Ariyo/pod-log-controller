<div align="center">
  <a href="https://www.fairwinds.com/">
    <img src="https://mms.businesswire.com/media/20190709005646/en/731887/23/FairwindsLogo.jpg" width="240"/>
  </a>
</div>

# FAIRWINDS POD CONTROLLER
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors)


The fairwinds pod controller is a kubernetes controller that performs custom actions to a kubernetes application which include:

```
   1. Listening to kubernetes for new pods
   2. Annotating these pods with a timestamp
   3. Logging the pods and timestamps to stdout
   4. Making deploys to a helm chart (podlog)
   5. Only responding to pods with a particular annotation
   6. Only responding to pods in namespaces with a particular annotation
   7. Implementing leader election
```


## Getting Started

  1. Follow the guidelines here to install kubernetes on your machine:

      https://kubernetes.io/docs/tasks/tools/

  2. Get a cluster and set it as your current configuration 
  
      https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/


### Installation

  *. To pull the docker image from a container registry, run the command below:

    ```docker pull lexmill99/operator-log:v3.2.0```

  *. To push a modified image to docker hub

    ```docker build -t <username>/<repository-name>:<tag> .```

    ```docker push <username>/<repository-name>:<tag>```


  *. To remove the chart from helm, run

    ``` helm uninstall podlog ```  

    ``` helm upgrade --install podlog podlog ```



## Contributing

Please read [CONTRIBUTING.md](https://www.dataschool.io/how-to-contribute-on-github/) for details on contributions and the process of submitting pull requests.

## Support & Contact

<div>
  <a  href="https://twitter.com/lay__kay" ><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"></a>
  <a href="https://t.me/lexmill99"><img src="https://img.shields.io/badge/Telegram-blue.svg"></a>
</div>


## License

This project is licensed under the MIT License; see the [LICENSE.md](LICENSE.md) file for details.

## Contributors ✨

Thanks to me ([emoji key](https://allcontributors.org/docs/en/)):



This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

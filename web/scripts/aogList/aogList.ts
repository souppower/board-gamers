import {ArrivalOfGame} from "../interfaces";
import IHttpService = angular.IHttpService;
import IWindowService = angular.IWindowService;
export class AogListController {
    aogs: ArrivalOfGame[];

    constructor(private $http: IHttpService, private $window:IWindowService) {
    }

    $routerOnActivate() {
        this.$http.get<ArrivalOfGame[]>("/api/v1/arrivalOfGames")
            .then((res) => {
                this.aogs = res.data;
            });
    }

    goTo(url: string) {
        this.$window.open(url);
    }
}

export const AogList = {
    name: "aogList",
    controller: AogListController,
    template: `
        <md-list layout-wrap>
            <md-subheader class="md-no-sticky">最新の入荷情報</md-subheader>
            <md-list-item class="md-3-line md-long-text" ng-repeat="aog in ::$ctrl.aogs" ng-click="$ctrl.goTo(aog.url)" aria-label="Go to Twitter">
                <div class="md-list-item-text" ng-cloak>
                    <h3>{{::aog.shop}}</h3>
                    <p><span ng-repeat="game in ::aog.games">{{::game}}{{$last ? "" : ", "}}</span></p>
                    <div am-time-ago="::aog.createdAt" am-format="YYYY-MM-DDThh:mm:ssZ"></div>
                </div>
            </md-list-item>
        </md-list>
    `
};

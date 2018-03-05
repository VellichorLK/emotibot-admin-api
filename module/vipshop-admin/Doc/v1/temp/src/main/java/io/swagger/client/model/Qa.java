/*
 * 机器人设置API
 * This is api document page for robot setting RestAPIs
 *
 * OpenAPI spec version: 1.0.0
 * Contact: danielwu@emotibot.com
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 * Do not edit the class manually.
 */


package io.swagger.client.model;

import java.util.Objects;
import com.google.gson.TypeAdapter;
import com.google.gson.annotations.JsonAdapter;
import com.google.gson.annotations.SerializedName;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * Qa
 */
@javax.annotation.Generated(value = "io.swagger.codegen.languages.java.JavaClientCodegen", date = "2018-03-01T18:34:36.180+08:00")
public class Qa {
@SerializedName("id")
  private Integer id = null;
  @SerializedName("main_questions")
  private List<String> mainQuestions = null;
  @SerializedName("related_questions")
  private List<String> relatedQuestions = null;
  @SerializedName("answers")
  private List<String> answers = null;
  @SerializedName("created_time")
  private String createdTime = null;
  
  public Qa id(Integer id) {
    this.id = id;
    return this;
  }

  
  /**
  * Get id
  * @return id
  **/
  @ApiModelProperty(value = "")
  public Integer getId() {
    return id;
  }
  public void setId(Integer id) {
    this.id = id;
  }
  
  public Qa mainQuestions(List<String> mainQuestions) {
    this.mainQuestions = mainQuestions;
    return this;
  }

  public Qa addMainQuestionsItem(String mainQuestionsItem) {
    
    if (this.mainQuestions == null) {
      this.mainQuestions = new ArrayList<String>();
    }
    
    this.mainQuestions.add(mainQuestionsItem);
    return this;
  }
  
  /**
  * Get mainQuestions
  * @return mainQuestions
  **/
  @ApiModelProperty(value = "")
  public List<String> getMainQuestions() {
    return mainQuestions;
  }
  public void setMainQuestions(List<String> mainQuestions) {
    this.mainQuestions = mainQuestions;
  }
  
  public Qa relatedQuestions(List<String> relatedQuestions) {
    this.relatedQuestions = relatedQuestions;
    return this;
  }

  public Qa addRelatedQuestionsItem(String relatedQuestionsItem) {
    
    if (this.relatedQuestions == null) {
      this.relatedQuestions = new ArrayList<String>();
    }
    
    this.relatedQuestions.add(relatedQuestionsItem);
    return this;
  }
  
  /**
  * Get relatedQuestions
  * @return relatedQuestions
  **/
  @ApiModelProperty(value = "")
  public List<String> getRelatedQuestions() {
    return relatedQuestions;
  }
  public void setRelatedQuestions(List<String> relatedQuestions) {
    this.relatedQuestions = relatedQuestions;
  }
  
  public Qa answers(List<String> answers) {
    this.answers = answers;
    return this;
  }

  public Qa addAnswersItem(String answersItem) {
    
    if (this.answers == null) {
      this.answers = new ArrayList<String>();
    }
    
    this.answers.add(answersItem);
    return this;
  }
  
  /**
  * Get answers
  * @return answers
  **/
  @ApiModelProperty(value = "")
  public List<String> getAnswers() {
    return answers;
  }
  public void setAnswers(List<String> answers) {
    this.answers = answers;
  }
  
  public Qa createdTime(String createdTime) {
    this.createdTime = createdTime;
    return this;
  }

  
  /**
  * Get createdTime
  * @return createdTime
  **/
  @ApiModelProperty(example = "2017-06-09T15:41:55+08:00", value = "")
  public String getCreatedTime() {
    return createdTime;
  }
  public void setCreatedTime(String createdTime) {
    this.createdTime = createdTime;
  }
  
  @Override
  public boolean equals(java.lang.Object o) {
    if (this == o) {
      return true;
    }
    if (o == null || getClass() != o.getClass()) {
      return false;
    }
    Qa qa = (Qa) o;
    return Objects.equals(this.id, qa.id) &&
        Objects.equals(this.mainQuestions, qa.mainQuestions) &&
        Objects.equals(this.relatedQuestions, qa.relatedQuestions) &&
        Objects.equals(this.answers, qa.answers) &&
        Objects.equals(this.createdTime, qa.createdTime);
  }

  @Override
  public int hashCode() {
    return Objects.hash(id, mainQuestions, relatedQuestions, answers, createdTime);
  }
  
  @Override
  public String toString() {
    StringBuilder sb = new StringBuilder();
    sb.append("class Qa {\n");
    
    sb.append("    id: ").append(toIndentedString(id)).append("\n");
    sb.append("    mainQuestions: ").append(toIndentedString(mainQuestions)).append("\n");
    sb.append("    relatedQuestions: ").append(toIndentedString(relatedQuestions)).append("\n");
    sb.append("    answers: ").append(toIndentedString(answers)).append("\n");
    sb.append("    createdTime: ").append(toIndentedString(createdTime)).append("\n");
    sb.append("}");
    return sb.toString();
  }

  /**
   * Convert the given object to string with each line indented by 4 spaces
   * (except the first line).
   */
  private String toIndentedString(java.lang.Object o) {
    if (o == null) {
      return "null";
    }
    return o.toString().replace("\n", "\n    ");
  }

  
}



